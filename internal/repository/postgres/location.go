package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/begenov/real-estate/internal/model"
	"strings"
	"unicode/utf8"
)

type ILocationRepo interface {
	// Country methods
	GetCountries(ctx context.Context) ([]*model.Country, error)
	CreateCountry(ctx context.Context, country *model.Country) error
	UpdateCountry(ctx context.Context, country *model.Country) error
	DeleteCountry(ctx context.Context, countryId int64) error
	GetCountryByID(ctx context.Context, id int64) (*model.Country, error)

	// Region methods
	GetRegions(ctx context.Context, filter *model.RegionFilter) ([]*model.Region, int, error)
	CreateRegion(ctx context.Context, region *model.Region) error
	UpdateRegion(ctx context.Context, region *model.Region) error
	DeleteRegion(ctx context.Context, regionId int64) error
	GetRegionByID(ctx context.Context, id int64) (*model.Region, error)

	// District methods
	GetDistricts(ctx context.Context, filter *model.DistrictFilter) ([]*model.District, int, error)
	CreateDistrict(ctx context.Context, district *model.District) error
	UpdateDistrict(ctx context.Context, district *model.District) error
	DeleteDistrict(ctx context.Context, districtId int64) error
	GetDistrictByID(ctx context.Context, id int64) (*model.District, error)
}

type LocationRepo struct {
	db *sql.DB
}

func NewLocationRepo(db *sql.DB) *LocationRepo {
	return &LocationRepo{
		db: db,
	}
}

//TODO country

func (r *LocationRepo) GetCountries(ctx context.Context) ([]*model.Country, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name_ru, name_en, name_tr, name_de, code, phone_code,  created_at, updated_at FROM public.country WHERE is_deleted = FALSE`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var countries []*model.Country
	for rows.Next() {
		var country model.Country
		if err := rows.Scan(&country.ID, &country.NameRu, &country.NameEn, &country.NameTr, &country.NameDe, &country.Code, &country.PhoneNumber, &country.CreatedAt, &country.UpdatedAt); err != nil {
			return nil, err
		}
		countries = append(countries, &country)
	}

	return countries, nil
}

func (r *LocationRepo) CreateCountry(ctx context.Context, country *model.Country) error {
	query := `INSERT INTO public.country (name_ru, name_en, name_tr, name_de, code, phone_code,  created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`

	_, err := r.db.ExecContext(ctx, query, country.NameRu, country.NameEn, country.NameTr, country.NameDe, country.Code, country.PhoneNumber)
	return err
}

func (r *LocationRepo) UpdateCountry(ctx context.Context, country *model.Country) error {
	query := `UPDATE public.country
              SET name_ru = $1, name_en = $2, name_tr = $3, name_de = $4, code = $5, phone_code = $6,  updated_at = NOW()
              WHERE id = $7 AND is_deleted = FALSE`

	_, err := r.db.ExecContext(ctx, query, country.NameRu, country.NameEn, country.NameTr, country.NameDe, country.Code, country.PhoneNumber, country.ID)
	return err
}

func (r *LocationRepo) DeleteCountry(ctx context.Context, countryId int64) error {
	query := `UPDATE public.country SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, countryId)
	return err
}

func (r *LocationRepo) GetCountryByID(ctx context.Context, id int64) (*model.Country, error) {
	var country model.Country
	query := `SELECT id, name_ru, name_en, name_tr, name_de, code, phone_code, created_at, updated_at 
              FROM public.country 
              WHERE id = $1 AND is_deleted = FALSE`

	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&country.ID, &country.NameRu, &country.NameEn, &country.NameTr, &country.NameDe, &country.Code, &country.PhoneNumber, &country.CreatedAt, &country.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &country, nil
}

//TODO region

func (r *LocationRepo) GetRegions(ctx context.Context, filter *model.RegionFilter) ([]*model.Region, int, error) {
	var regions []*model.Region
	var count int

	var params []interface{}
	var queryBuilder strings.Builder

	queryBuilder.WriteString(`SELECT id, name_ru, name_en, name_tr, name_de, country_id, created_at, updated_at 
              FROM public.region 
              WHERE is_deleted = FALSE`)

	if filter != nil {
		if filter.CountryId != nil && *filter.CountryId > 0 {
			queryBuilder.WriteString(fmt.Sprintf(` AND country_id = $%d`, len(params)+1))
			params = append(params, *filter.CountryId)
		}
		if filter.Name != nil && utf8.RuneCountInString(*filter.Name) > 0 {
			queryBuilder.WriteString(fmt.Sprintf(` AND (name_ru ILIKE $%d OR name_en ILIKE $%d OR name_tr ILIKE $%d OR name_de ILIKE %d)`, len(params)+1, len(params)+1, len(params)+1, len(params)+1))
			params = append(params, "%"+*filter.Name+"%")
		}

		queryBuilder.WriteString(fmt.Sprintf(` LIMIT $%d OFFSET $%d `, len(params)+1, len(params)+2))

		params = append(params, filter.Rows, filter.Page*filter.Rows)
	}

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), params...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var region model.Region
		region.Country = &model.Country{}
		if err := rows.Scan(&region.ID, &region.NameRu, &region.NameEn, &region.NameTr, &region.NameDe, &region.Country.ID,
			&region.CreatedAt, &region.UpdatedAt); err != nil {
			return nil, 0, err
		}
		regions = append(regions, &region)
	}

	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM public.region WHERE is_deleted = FALSE`).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return regions, count, nil
}

func (r *LocationRepo) CreateRegion(ctx context.Context, region *model.Region) error {
	query := `INSERT INTO public.region (name_ru, name_en, name_tr, name_de, country_id, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, now(), now())`

	_, err := r.db.ExecContext(ctx, query,
		region.NameRu,
		region.NameEn,
		region.NameTr,
		region.NameDe,
		region.Country.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to create region: %w", err)
	}

	return nil
}

func (r *LocationRepo) UpdateRegion(ctx context.Context, region *model.Region) error {
	query := `UPDATE public.region 
			  SET name_ru = $1, name_en = $2, name_tr = $3, name_de = $4, country_id = $5, updated_at = now() 
			  WHERE id = $6 AND is_deleted = FALSE`

	_, err := r.db.ExecContext(ctx, query,
		region.NameRu,
		region.NameEn,
		region.NameTr,
		region.NameDe,
		region.Country.ID,
		region.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update region: %w", err)
	}

	return nil
}

func (r *LocationRepo) DeleteRegion(ctx context.Context, regionId int64) error {
	query := `UPDATE public.region
			  SET is_deleted = TRUE, updated_at = now()
			  WHERE id = $1 AND is_deleted = FALSE`

	_, err := r.db.ExecContext(ctx, query, regionId)
	if err != nil {
		return fmt.Errorf("failed to delete region: %w", err)
	}

	return nil
}

func (r *LocationRepo) GetRegionByID(ctx context.Context, id int64) (*model.Region, error) {
	query := `SELECT id, name_ru, name_en, name_tr, name_de, country_id, created_at, updated_at
			  FROM public.region
			  WHERE id = $1 AND is_deleted = FALSE`

	var region model.Region
	region.Country = &model.Country{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&region.ID, &region.NameRu, &region.NameEn, &region.NameTr, &region.NameDe,
		&region.Country.ID,
		&region.CreatedAt, &region.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("region with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get region by id: %w", err)
	}

	return &region, nil
}

//TODO district

func (r *LocationRepo) GetDistricts(ctx context.Context, filter *model.DistrictFilter) ([]*model.District, int, error) {
	var args []interface{}
	query := `SELECT id, name_ru, name_en, name_tr, name_de, region_id, created_at, updated_at
			  FROM public.district WHERE is_deleted = FALSE`
	countQuery := `SELECT COUNT(*) FROM public.district WHERE is_deleted = FALSE`

	if filter != nil {
		if filter.RegionId != nil && *filter.RegionId > 0 {
			args = append(args, filter.RegionId)
			query += ` AND region_id = $` + fmt.Sprintf("%d", len(args))
			countQuery += ` AND region_id = $` + fmt.Sprintf("%d", len(args))
		}
		if filter.Name != nil && utf8.RuneCountInString(*filter.Name) > 0 {
			args = append(args, "%"+*filter.Name+"%")
			query += ` AND (name_ru ILIKE $` + fmt.Sprintf("%d", len(args)) +
				` OR name_en ILIKE $` + fmt.Sprintf("%d", len(args)) +
				` OR name_tr ILIKE $` + fmt.Sprintf("%d", len(args)) +
				` OR name_de ILIKE $` + fmt.Sprintf("%d", len(args)) + `)`
			countQuery += ` AND (name_ru ILIKE $` + fmt.Sprintf("%d", len(args)) +
				` OR name_en ILIKE $` + fmt.Sprintf("%d", len(args)) +
				` OR name_tr ILIKE $` + fmt.Sprintf("%d", len(args)) +
				` OR name_de ILIKE $` + fmt.Sprintf("%d", len(args)) + `)`
		}

	}

	var count int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	if filter != nil {
		args = append(args, filter.Rows)
		query += ` LIMIT $` + fmt.Sprintf("%d", len(args))

		args = append(args, filter.Page*filter.Rows)
		query += ` OFFSET $` + fmt.Sprintf("%d", len(args))
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var districts []*model.District
	for rows.Next() {
		var district model.District
		district.Region = &model.Region{}
		err := rows.Scan(
			&district.ID,
			&district.NameRu,
			&district.NameEn,
			&district.NameTr,
			&district.NameDe,
			&district.Region.ID,
			&district.CreatedAt,
			&district.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		districts = append(districts, &district)
	}

	return districts, count, nil
}

func (r *LocationRepo) CreateDistrict(ctx context.Context, district *model.District) error {
	query := `INSERT INTO public.district (name_ru, name_en, name_tr, name_de,  region_id, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5,  now(), now())`

	_, err := r.db.ExecContext(ctx, query,
		district.NameRu,
		district.NameEn,
		district.NameTr,
		district.NameDe,
		district.Region.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to create district: %w", err)
	}
	return nil
}

func (r *LocationRepo) UpdateDistrict(ctx context.Context, district *model.District) error {
	query := `UPDATE public.district
			  SET name_ru = $1, name_en = $2, name_tr = $3, name_de = $4, region_id = $5, updated_at = now()
			  WHERE id = $6 AND is_deleted = FALSE`

	_, err := r.db.ExecContext(ctx, query,
		district.NameRu,
		district.NameEn,
		district.NameTr,
		district.NameDe,
		district.Region.ID,
		district.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update district: %w", err)
	}
	return nil
}

func (r *LocationRepo) DeleteDistrict(ctx context.Context, districtId int64) error {
	query := `UPDATE public.district
			  SET is_deleted = TRUE, updated_at = now()
			  WHERE id = $1 AND is_deleted = FALSE`

	_, err := r.db.ExecContext(ctx, query, districtId)
	if err != nil {
		return fmt.Errorf("failed to delete district: %w", err)
	}
	return nil
}

func (r *LocationRepo) GetDistrictByID(ctx context.Context, id int64) (*model.District, error) {
	query := `SELECT id, name_ru, name_en, name_tr, name_de, region_id, created_at, updated_at
			  FROM public.district
			  WHERE id = $1 AND is_deleted = FALSE`

	var district model.District
	district.Region = &model.Region{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&district.ID,
		&district.NameRu,
		&district.NameEn,
		&district.NameTr,
		&district.NameDe,
		&district.Region.ID,
		&district.CreatedAt,
		&district.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get district by id: %w", err)
	}
	return &district, nil
}
