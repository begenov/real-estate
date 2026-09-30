package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/pkg/helper"

	"github.com/lib/pq"
	"github.com/minio/minio-go/v7"
)

type IRealEstateRepo interface {
	GetRealEstates(ctx context.Context, filter model.RealEstateFilter) (model.RealEstates, int, error)
	CreateRealEstate(ctx context.Context, tx *sql.Tx, realEstate *model.RealEstateInput) (int64, error)
	UpsertRealEstateURLs(ctx context.Context, tx *sql.Tx, realEstateId int64, photos []model.RealEstatePhoto) error
	SetStatus(ctx context.Context, tx *sql.Tx, realEstateId, userId int64, status int) error
	GetRealEstateURL(ctx context.Context, realEstateId int64) (*model.RealEstatePhoto, error)
	GetRealEstate(ctx context.Context, id int64) (*model.RealEstate, error)
	GetRealEstateURLs(ctx context.Context, realEstateId int64) ([]model.RealEstatePhoto, error)
	GetRealEstateURLsByIDs(ctx context.Context, realEstateId int64, photoIDs []int64) ([]model.RealEstatePhoto, error)
	GetRealEstateTranslationsByIDs(ctx context.Context, realEstateIDs []int64) (map[int64][]model.RealEstateTranslation, error)
	GetRealEstateLastHistories(ctx context.Context, realEstateIDs []int64) (map[int64]*model.RealEstateHistory, error)

	GetRealEstateLastHistory(ctx context.Context, realEstateId int64) (*model.RealEstateHistory, error)
	GetTotalRealEstates(ctx context.Context, filter model.RealEstateFilter) (int, error)

	UpdateRealEstate(ctx context.Context, tx *sql.Tx, realEstate *model.RealEstateInput) error
	DeleteSelectedRealEstatePhotos(ctx context.Context, tx *sql.Tx, realEstateId int64, photoIDs []int64) error

	UpsertRealEstateTranslations(ctx context.Context, tx *sql.Tx, realEstateId int64, translations []model.RealEstateTranslationInput) error
	DeleteRealEstate(ctx context.Context, tx *sql.Tx, id int64) error
	DeleteRealEstatePhotos(ctx context.Context, tx *sql.Tx, realEstateId int64) error

	GetRealEstateTranslations(ctx context.Context, realEstateId int64) ([]model.RealEstateTranslation, error)

	AddAmenity(ctx context.Context, tx *sql.Tx, realEstateID, amenityID int64) error
	RemoveAmenity(ctx context.Context, tx *sql.Tx, realEstateID, amenityID int64) error
	GetAmenitiesByRealEstateID(ctx context.Context, realEstateID int64) ([]int64, error)
}
type RealEstateRepo struct {
	db          *sql.DB
	minioClient *minio.Client
}

func NewRealEstateRepo(db *sql.DB, minioClient *minio.Client) IRealEstateRepo {
	return &RealEstateRepo{
		db:          db,
		minioClient: minioClient,
	}
}

func (r *RealEstateRepo) GetRealEstate(ctx context.Context, id int64) (*model.RealEstate, error) {
	query := `
		SELECT re.id, re.price, re.area, re.rooms, re.region_id,
		       re."location", re.completion_date, re.latitude, re.longitude,
		       u.id, u.first_name, u.last_name, u.middle_name, u.phone, u.email,
		       re_type.id, re_type."name", 
		       re_purpose.id, re_purpose."name", 
		       re.address, re.apartment,
		       re.floor, re.total_floors, re.parking_available, re.built_year,
		       re.price_per_square_meter, re.has_balcony, re.distance_to_sea, re.district_id
		FROM public.real_estate re
		JOIN public."user" u ON u.id = re.manager_id
		LEFT JOIN public.real_estate_type re_type ON re_type.id = re.type
		LEFT JOIN public.real_estate_purpose re_purpose ON re_purpose.id = re.purpose_id
		WHERE re.id = $1 AND re.is_deleted = false
		LIMIT 1;
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var realEstate model.RealEstate

	realEstate.Type = &model.Type{}
	realEstate.Purpose = &model.RealEstatePurpose{}
	realEstate.Manager = model.User{}

	err := row.Scan(
		&realEstate.ID, &realEstate.Price, &realEstate.Area, &realEstate.Rooms, &realEstate.Region.ID,
		&realEstate.Location, &realEstate.CompletionDate, &realEstate.Latitude, &realEstate.Longitude,
		&realEstate.Manager.ID, &realEstate.Manager.FirstName, &realEstate.Manager.LastName,
		&realEstate.Manager.MiddleName, &realEstate.Manager.Phone, &realEstate.Manager.Email,
		&realEstate.Type.ID, &realEstate.Type.Name,
		&realEstate.Purpose.ID, &realEstate.Purpose.Name,
		&realEstate.Address, &realEstate.Apartment,
		&realEstate.Floor, &realEstate.TotalFloors, &realEstate.ParkingAvailable, &realEstate.BuiltYear,
		&realEstate.PricePerSquareMeter, &realEstate.HasBalcony, &realEstate.DistanceToSea, &realEstate.District.ID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &realEstate, nil
}

func (r *RealEstateRepo) GetRealEstates(ctx context.Context, filter model.RealEstateFilter) (model.RealEstates, int, error) {
	var queryBuilder strings.Builder
	var params []interface{}

	statement := `SELECT re.id, re.price, re.area, re.rooms,
		re.region_id, r.name_ru, r.name_en, r.name_tr, r.name_de,
		re.completion_date,
		re.latitude, re.longitude, re.manager_id,
		u.first_name, u.last_name, u.middle_name, u.phone, u.email,
		re_type.id, re_type."name", 
		re_purpose.id, re_purpose."name", 
		re.address, re.apartment,
		re.floor, re.total_floors, re.parking_available, re.built_year,
		re.price_per_square_meter, re.has_balcony, re.distance_to_sea,
		re.district_id, d.name_ru, d.name_en, d.name_tr, d.name_de,
		photo.photo_id, photo.photo_url
	`

	queryBuilder.WriteString(`
		FROM public.real_estate re
		JOIN public."user" u ON u.id = re.manager_id
		JOIN public.real_estate_history reh ON reh.real_estate_id = re.id
		LEFT JOIN public.real_estate_type re_type ON re_type.id = re.type
		LEFT JOIN public.real_estate_purpose re_purpose ON re_purpose.id = re.purpose_id
		LEFT JOIN public.collection_item ci on ci.real_estate_id = re.id 
		LEFT JOIN public.region r ON r.id = re.region_id
		LEFT JOIN public.district d ON d.id = re.district_id
		LEFT JOIN LATERAL (
			SELECT f.id as photo_id, f.url as photo_url
			FROM public.real_estate_url reu
			JOIN public.files f ON f.id = reu.files_id
			WHERE reu.real_estate_id = re.id
			ORDER BY reu.order_num ASC
			LIMIT 1
		) photo ON true
		WHERE re.is_deleted = false
	`)

	if filter.PriceMin != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.price >= $%d", len(params)+1))
		params = append(params, *filter.PriceMin)
	}

	if filter.PriceMax != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.price <= $%d", len(params)+1))
		params = append(params, *filter.PriceMax)
	}

	if filter.RegionID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.region_id = $%d", len(params)+1))
		params = append(params, *filter.RegionID)
	}

	if len(filter.Rooms) > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.rooms IN (%s)", helper.GeneratePlaceholders(len(filter.Rooms), len(params)+1)))
		for i := range filter.Rooms {
			params = append(params, filter.Rooms[i])
		}
	}

	if filter.Purpose != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.purpose_id in (%s)", helper.GeneratePlaceholders(len(filter.Purpose), len(params)+1)))
		for i := range filter.Purpose {
			params = append(params, filter.Purpose[i])
		}
	}

	if len(filter.Status) > 0 {

		queryBuilder.WriteString(` AND
		reh.id = (
    	SELECT MAX(reh2.id)
    	FROM public.real_estate_history reh2
    	WHERE reh2.real_estate_id = re.id
  		)
`)

		queryBuilder.WriteString(fmt.Sprintf(" AND reh.status_id IN (%s)", helper.GeneratePlaceholders(len(filter.Status), len(params)+1)))
		for i := range filter.Status {
			params = append(params, filter.Status[i])
		}
	}

	if filter.CollectionId != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND ci.collection_id = $%d", len(params)+1))
		params = append(params, *filter.CollectionId)
	}

	if filter.DistrictID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.district_id = $%d", len(params)+1))
		params = append(params, *filter.DistrictID)
	}

	if len(filter.Type) > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.type in (%s)", helper.GeneratePlaceholders(len(filter.Type), len(params)+1)))
		for i := range filter.Type {
			params = append(params, filter.Type[i])
		}
	}

	if filter.ID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.id = $%d", len(params)+1))
		params = append(params, *filter.ID)
	}

	var totalCount int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(distinct re.id) `+queryBuilder.String(), params...).Scan(&totalCount)
	if err != nil {
		logger.Error(err)
		return nil, 0, err
	}

	if len(filter.SortBy) > 0 {
		var sortClauses []string
		validSortColumns := map[string]bool{"price": true, "area": true, "rooms": true}

		for i, sortField := range filter.SortBy {
			if validSortColumns[sortField] {
				order := "ASC"
				if i < len(filter.SortOrder) && filter.SortOrder[i] == "desc" {
					order = "DESC"
				}
				sortClauses = append(sortClauses, fmt.Sprintf("%s %s", sortField, order))
			}
		}

		if len(sortClauses) > 0 {
			queryBuilder.WriteString(" ORDER BY " + strings.Join(sortClauses, ", "))
		}
	}

	firstIndex := (filter.Page - 1) * filter.Rows
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Rows, firstIndex))

	rows, err := r.db.QueryContext(ctx, statement+queryBuilder.String(), params...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, model.ErrNotFound
		}

		logger.Error(err)
		return nil, 0, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var estates = make(model.RealEstates, 0, filter.Rows)
	for rows.Next() {
		var estate model.RealEstate
		var regionNameRu, regionNameEn, regionNameTr, regionNameDe sql.NullString
		var districtNameRu, districtNameEn, districtNameTr, districtNameDe sql.NullString
		var photoID sql.NullInt64
		var photoURL sql.NullString

		estate.Type = &model.Type{}
		estate.Purpose = &model.RealEstatePurpose{}
		estate.Manager = model.User{}

		err := rows.Scan(
			&estate.ID, &estate.Price, &estate.Area, &estate.Rooms, &estate.Region.ID,
			&regionNameRu, &regionNameEn, &regionNameTr, &regionNameDe,
			&estate.CompletionDate,
			&estate.Latitude, &estate.Longitude, &estate.Manager.ID,
			&estate.Manager.FirstName, &estate.Manager.LastName,
			&estate.Manager.MiddleName, &estate.Manager.Phone,
			&estate.Manager.Email,
			&estate.Type.ID, &estate.Type.Name,
			&estate.Purpose.ID, &estate.Purpose.Name,
			&estate.Address, &estate.Apartment,
			&estate.Floor, &estate.TotalFloors, &estate.ParkingAvailable, &estate.BuiltYear,
			&estate.PricePerSquareMeter, &estate.HasBalcony, &estate.DistanceToSea,
			&estate.District.ID, &districtNameRu, &districtNameEn, &districtNameTr, &districtNameDe,
			&photoID, &photoURL,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, 0, model.ErrNotFound
			}

			return nil, 0, err
		}

		if regionNameRu.Valid {
			estate.Region.NameRu = regionNameRu.String
		}
		if regionNameEn.Valid {
			estate.Region.NameEn = regionNameEn.String
		}
		if regionNameTr.Valid {
			estate.Region.NameTr = regionNameTr.String
		}
		if regionNameDe.Valid {
			estate.Region.NameDe = regionNameDe.String
		}

		if districtNameRu.Valid {
			estate.District.NameRu = districtNameRu.String
		}
		if districtNameEn.Valid {
			estate.District.NameEn = districtNameEn.String
		}
		if districtNameTr.Valid {
			estate.District.NameTr = districtNameTr.String
		}
		if districtNameDe.Valid {
			estate.District.NameDe = districtNameDe.String
		}

		if photoID.Valid && photoURL.Valid {
			estate.Photos = append(estate.Photos, model.RealEstatePhoto{
				ID:  photoID.Int64,
				URL: photoURL.String,
			})
		}

		estates = append(estates, estate)
	}

	return estates, totalCount, nil
}

func (r *RealEstateRepo) GetRealEstateTranslationsByIDs(ctx context.Context, realEstateIDs []int64) (map[int64][]model.RealEstateTranslation, error) {
	if len(realEstateIDs) == 0 {
		return map[int64][]model.RealEstateTranslation{}, nil
	}

	query := `
		SELECT ret.real_estate_id, ret.id, ret.description, ret."name",
		       ret.lang_id, lang.code, ret.summary_title, ret.terms_and_conditions
		FROM real_estate_translations ret
		JOIN language lang ON lang.id = ret.lang_id
		WHERE ret.real_estate_id = ANY($1);
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(realEstateIDs))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	result := make(map[int64][]model.RealEstateTranslation)
	for rows.Next() {
		var estateID int64
		var t model.RealEstateTranslation
		if err := rows.Scan(&estateID, &t.ID, &t.Description, &t.Name, &t.Language.ID, &t.Language.Code, &t.SummaryTitle, &t.TermsAndConditions); err != nil {
			return nil, err
		}
		result[estateID] = append(result[estateID], t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *RealEstateRepo) GetRealEstateLastHistories(ctx context.Context, realEstateIDs []int64) (map[int64]*model.RealEstateHistory, error) {
	if len(realEstateIDs) == 0 {
		return map[int64]*model.RealEstateHistory{}, nil
	}

	query := `
		SELECT DISTINCT ON (reh.real_estate_id)
		       reh.real_estate_id, reh.id, reh.status_id, s.name, reh.created_at, reh.manager_id,
		       u.first_name, u.last_name, u.middle_name, u.phone, u.email
		FROM public.real_estate_history reh
		JOIN "user" u ON u.id = reh.manager_id
		JOIN status s ON s.id = reh.status_id
		WHERE reh.real_estate_id = ANY($1)
		ORDER BY reh.real_estate_id, reh.id DESC;
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(realEstateIDs))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	result := make(map[int64]*model.RealEstateHistory)
	for rows.Next() {
		var estateID int64
		var history model.RealEstateHistory
		if err := rows.Scan(
			&estateID,
			&history.ID,
			&history.Status.ID,
			&history.Status.Name,
			&history.CreatedAt,
			&history.Manager.ID,
			&history.Manager.FirstName,
			&history.Manager.LastName,
			&history.Manager.MiddleName,
			&history.Manager.Phone,
			&history.Manager.Email,
		); err != nil {
			return nil, err
		}
		history.RealEstateID = int(estateID)
		result[estateID] = &history
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *RealEstateRepo) GetRealEstateLastHistory(ctx context.Context, realEstateId int64) (*model.RealEstateHistory, error) {
	var history model.RealEstateHistory

	query := `
		SELECT reh.id, reh.real_estate_id, reh.status_id, s.name, reh.created_at, reh.manager_id, 
		u.first_name, u.last_name, u.middle_name, u.phone, u.email
		FROM public.real_estate_history reh
		join "user" u on u.id = reh.manager_id 
		join status s on s.id = reh.status_id 
		WHERE real_estate_id = $1
		ORDER BY id DESC
		LIMIT 1;
	`

	err := r.db.QueryRowContext(ctx, query, realEstateId).Scan(
		&history.ID,
		&history.RealEstateID,
		&history.Status.ID,
		&history.Status.Name,
		&history.CreatedAt,
		&history.Manager.ID,
		&history.Manager.FirstName,
		&history.Manager.LastName,
		&history.Manager.MiddleName,
		&history.Manager.Phone,
		&history.Manager.Email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &history, nil
}

func (r *RealEstateRepo) GetRealEstateURLs(ctx context.Context, realEstateId int64) ([]model.RealEstatePhoto, error) {
	query := `
		select  f.id, f.url 
		from files f 
		inner join real_estate_url reu ON reu.files_id = f.id 
		where reu.real_estate_id = $1
		order by reu.order_num asc
	`

	rows, err := r.db.QueryContext(ctx, query, realEstateId)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var urls []model.RealEstatePhoto
	for rows.Next() {
		var url model.RealEstatePhoto
		if err := rows.Scan(&url.ID, &url.URL); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, model.ErrNotFound
			}

			return nil, err
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (r *RealEstateRepo) GetRealEstateURLsByIDs(ctx context.Context, realEstateId int64, photoIDs []int64) ([]model.RealEstatePhoto, error) {
	if len(photoIDs) == 0 {
		return []model.RealEstatePhoto{}, nil
	}

	query := `
		select f.id, f.url
		from files f
		inner join real_estate_url reu ON reu.files_id = f.id
		where reu.real_estate_id = $1 and f.id = ANY($2)
	`

	rows, err := r.db.QueryContext(ctx, query, realEstateId, pq.Array(photoIDs))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var urls []model.RealEstatePhoto
	for rows.Next() {
		var url model.RealEstatePhoto
		if err := rows.Scan(&url.ID, &url.URL); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, model.ErrNotFound
			}
			return nil, err
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (r *RealEstateRepo) GetRealEstateURL(ctx context.Context, realEstateId int64) (*model.RealEstatePhoto, error) {
	query := `
		select  f.id, f.url 
		from files f 
		inner join real_estate_url reu ON reu.files_id = f.id 
		where reu.real_estate_id = $1
		order by reu.order_num asc
		limit 1
		`

	var url model.RealEstatePhoto

	err := r.db.QueryRowContext(ctx, query, realEstateId).Scan(&url.ID, &url.URL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, err
	}

	return &url, nil
}

func (r *RealEstateRepo) CreateRealEstate(ctx context.Context, tx *sql.Tx, realEstate *model.RealEstateInput) (int64, error) {

	query := `
		INSERT INTO public.real_estate (
			price, area, rooms, region_id, "type", purpose_id, latitude, longitude, completion_date, manager_id, location, address, apartment,
			floor, total_floors, parking_available, built_year, price_per_square_meter, has_balcony, distance_to_sea, district_id
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		)
		RETURNING id
	`

	var realEstateID int64
	err := tx.QueryRowContext(ctx, query,
		realEstate.Price, realEstate.Area, realEstate.Rooms, realEstate.RegionID, realEstate.Type.ID,
		realEstate.Purpose.ID, realEstate.Latitude, realEstate.Longitude, realEstate.CompletionDate,
		realEstate.OwnerId, realEstate.Location, realEstate.Address, realEstate.Apartment,
		realEstate.Floor, realEstate.TotalFloors, realEstate.ParkingAvailable, realEstate.BuiltYear,
		realEstate.PricePerSquareMeter, realEstate.HasBalcony, realEstate.DistanceToSea, realEstate.DistrictID,
	).Scan(&realEstateID)

	if err != nil {
		return 0, err
	}

	return realEstateID, nil
}

func (r *RealEstateRepo) UpsertRealEstateURLs(ctx context.Context, tx *sql.Tx, realEstateId int64, photos []model.RealEstatePhoto) error {
	query := `
		INSERT INTO public.real_estate_url (real_estate_id, files_id, order_num)
		VALUES ($1, $2, $3)
		ON CONFLICT (real_estate_id, files_id) DO UPDATE SET
			order_num = EXCLUDED.order_num;
	`

	for index, photo := range photos {
		_, err := tx.ExecContext(ctx, query, realEstateId, photo.ID, index+1)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RealEstateRepo) SetStatus(ctx context.Context, tx *sql.Tx, realEstateId, userId int64, status int) error {
	query := `
		INSERT INTO public.real_estate_history (real_estate_id, status_id, created_at, manager_id)
		VALUES ($1, $2, $3, $4)
	`

	_, err := tx.ExecContext(ctx, query, realEstateId, status, time.Now(), userId)
	if err != nil {
		return fmt.Errorf("ошибка при установке статуса: %w", err)
	}

	return nil
}

func (r *RealEstateRepo) DeleteRealEstatePhoto(tx *sql.Tx, photoId, realEstateId int64) error {
	query := `
		DELETE FROM public.real_estate_url 
		WHERE real_estate_id = $1 AND files_id = $2
	`
	_, err := tx.Exec(query, realEstateId, photoId)
	return err
}

func (r *RealEstateRepo) GetTotalRealEstates(ctx context.Context, filter model.RealEstateFilter) (int, error) {
	var queryBuilder strings.Builder
	var params []interface{}

	queryBuilder.WriteString(`
		SELECT COUNT(DISTINCT re.id)
		FROM public.real_estate re
		JOIN public."user" u ON u.id = re.manager_id
		JOIN public.real_estate_history reh ON reh.real_estate_id = re.id
		LEFT JOIN public.district d ON d.id = re.district_id
		WHERE 1=1
	`)

	if filter.PriceMin != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.price >= $%d", len(params)+1))
		params = append(params, *filter.PriceMin)
	}

	if filter.PriceMax != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.price <= $%d", len(params)+1))
		params = append(params, *filter.PriceMax)
	}

	if filter.RegionID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.region_id = $%d", len(params)+1))
		params = append(params, *filter.RegionID)
	}

	if len(filter.Rooms) > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.rooms IN (%s)", helper.GeneratePlaceholders(len(filter.Rooms), len(params)+1)))
		for _, room := range filter.Rooms {
			params = append(params, room)
		}
	}

	if filter.Purpose != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.purpose_id IN (%s)", helper.GeneratePlaceholders(len(filter.Purpose), len(params)+1)))
		for _, purpose := range filter.Purpose {
			params = append(params, purpose)
		}
	}

	if len(filter.Status) > 0 {
		queryBuilder.WriteString(` AND reh.id = (
			SELECT MAX(reh2.id)
			FROM public.real_estate_history reh2
			WHERE reh2.real_estate_id = re.id
		)`)
		queryBuilder.WriteString(fmt.Sprintf(" AND reh.status_id IN (%s)", helper.GeneratePlaceholders(len(filter.Status), len(params)+1)))
		for _, status := range filter.Status {
			params = append(params, status)
		}
	}

	if filter.DistrictID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND re.district_id = $%d", len(params)+1))
		params = append(params, *filter.DistrictID)
	}

	var totalCount int
	err := r.db.QueryRowContext(ctx, queryBuilder.String(), params...).Scan(&totalCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, model.ErrNotFound
		}

		logger.Error(err)
		return 0, err
	}

	return totalCount, nil
}

func (r *RealEstateRepo) UpdateRealEstate(ctx context.Context, tx *sql.Tx, realEstate *model.RealEstateInput) error {
	query := `
		UPDATE public.real_estate 
		SET price = $1, area = $2, rooms = $3, region_id = $4, "type" = $5, 
		    purpose_id = $6, latitude = $7, longitude = $8, completion_date = $9,
		    floor = $10, total_floors = $11, parking_available = $12, built_year = $13,
		    price_per_square_meter = $14, has_balcony = $15, distance_to_sea = $16, district_id = $17, address = $18
		WHERE id = $19
	`

	_, err := tx.ExecContext(ctx, query,
		realEstate.Price, realEstate.Area, realEstate.Rooms, realEstate.RegionID, realEstate.Type.ID,
		realEstate.Purpose.ID, realEstate.Latitude, realEstate.Longitude, realEstate.CompletionDate,
		realEstate.Floor, realEstate.TotalFloors, realEstate.ParkingAvailable, realEstate.BuiltYear,
		realEstate.PricePerSquareMeter, realEstate.HasBalcony, realEstate.DistanceToSea, realEstate.DistrictID,
		realEstate.Address,
		realEstate.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *RealEstateRepo) DeleteSelectedRealEstatePhotos(ctx context.Context, tx *sql.Tx, realEstateId int64, photoIDs []int64) error {
	query := `
		DELETE FROM public.real_estate_url 
		WHERE real_estate_id = $1 AND files_id = ANY($2)
	`
	_, err := tx.ExecContext(ctx, query, realEstateId, pq.Array(photoIDs))
	return err
}

func (r *RealEstateRepo) UpsertRealEstateTranslations(ctx context.Context, tx *sql.Tx, realEstateId int64, translations []model.RealEstateTranslationInput) error {
	query := `
		INSERT INTO real_estate_translations (real_estate_id, description, "name", lang_id, summary_title, terms_and_conditions)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (real_estate_id, lang_id) 
		DO UPDATE SET 
			description = EXCLUDED.description, 
			"name" = EXCLUDED.name, 
			summary_title = EXCLUDED.summary_title,
			terms_and_conditions = EXCLUDED.terms_and_conditions;
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()

	for _, t := range translations {
		_, err := stmt.ExecContext(ctx, realEstateId, t.Description, t.Title, t.Language.ID, t.SummaryTitle, t.TermsAndConditions)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RealEstateRepo) DeleteRealEstate(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `
		UPDATE public.real_estate 
		SET is_deleted = TRUE
		WHERE id = $1 AND is_deleted = FALSE;
	`

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *RealEstateRepo) DeleteRealEstatePhotos(ctx context.Context, tx *sql.Tx, realEstateId int64) error {
	query := `
		DELETE FROM public.real_estate_url 
		WHERE real_estate_id = $1
	`
	_, err := tx.ExecContext(ctx, query, realEstateId)
	return err
}

func (r *RealEstateRepo) GetRealEstateTranslations(ctx context.Context, realEstateId int64) ([]model.RealEstateTranslation, error) {
	var translations []model.RealEstateTranslation

	query := `
		SELECT ret.id, ret.real_estate_id, ret.description, ret."name" AS translation_name, ret.lang_id, lang.code, ret.summary_title, ret.terms_and_conditions
		FROM real_estate_translations ret
		JOIN language lang ON lang.id = ret.lang_id
		WHERE ret.real_estate_id = $1;
	`

	rows, err := r.db.QueryContext(ctx, query, realEstateId)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var t model.RealEstateTranslation
		err := rows.Scan(&t.ID, &t.RealEstateID, &t.Description, &t.Name, &t.Language.ID, &t.Language.Code, &t.SummaryTitle, &t.TermsAndConditions)
		if err != nil {
			return nil, err
		}
		translations = append(translations, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return translations, nil
}

func (r *RealEstateRepo) AddAmenity(ctx context.Context, tx *sql.Tx, realEstateID, amenityID int64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO public.real_estate_amenities (real_estate_id, amenity_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, realEstateID, amenityID)
	return err
}

func (r *RealEstateRepo) RemoveAmenity(ctx context.Context, tx *sql.Tx, realEstateID, amenityID int64) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM public.real_estate_amenities
		WHERE real_estate_id = $1 AND amenity_id = $2
	`, realEstateID, amenityID)
	return err
}

func (r *RealEstateRepo) GetAmenitiesByRealEstateID(ctx context.Context, realEstateID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT amenity_id
		FROM public.real_estate_amenities
		WHERE real_estate_id = $1
	`, realEstateID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var amenities []int64
	for rows.Next() {
		var amenityID int64
		if err := rows.Scan(&amenityID); err != nil {
			return nil, err
		}
		amenities = append(amenities, amenityID)
	}

	return amenities, rows.Err()
}
