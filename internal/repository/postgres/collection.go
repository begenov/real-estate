package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/pkg/helper"
	"strings"

	"github.com/lib/pq"
)

type ICollectionRepo interface {
	CreateCollection(ctx context.Context, tx *sql.Tx, collection *model.CollectionInput) (int64, error)
	SetStatus(ctx context.Context, tx *sql.Tx, collectionID, userId int64, statusID int) error
	UpsertCollectionItem(ctx context.Context, tx *sql.Tx, collectionID, realEstateID int64) error
	GetCollection(ctx context.Context, filter *model.CollectionFilter) (*model.Collection, error)
	UpdateCollection(ctx context.Context, collection *model.CollectionInput) error
	DeleteCollectionItem(ctx context.Context, tx *sql.Tx, collectionID, realEstateID int64) error
	GetCollections(ctx context.Context, filter *model.CollectionFilter) (model.Collections, int, error)
	DeleteCollectionItems(ctx context.Context, tx *sql.Tx, collectionID int64) error
	DeleteCollection(ctx context.Context, tx *sql.Tx, collectionID int64) error
	GetCollectionLastHistory(ctx context.Context, collectionId int64) (*model.CollectionHistory, error)
	GetCollectionByUUID(ctx context.Context, uuid string) (*model.Collection, error)
	SetCollectionUUID(ctx context.Context, collectionId int64, uuid string) error
	GetLatestHistoryByCollectionID(ctx context.Context, collectionID int64) (*model.CollectionHistory, error)

	InsertTranslation(ctx context.Context, tx *sql.Tx, collectionID int64, langID int, name, description string) error
	GetCollectionTranslations(ctx context.Context, collectionID int64) ([]model.CollectionTranslation, error)
}
type CollectionRepo struct {
	db *sql.DB
}

func NewCollectionRepo(db *sql.DB) ICollectionRepo {
	return &CollectionRepo{
		db: db,
	}
}

func (r *CollectionRepo) CreateCollection(ctx context.Context, tx *sql.Tx, collection *model.CollectionInput) (int64, error) {
	query := `INSERT INTO collection (name, description, manager_id, photo_url, is_temporary) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id int64
	err := tx.QueryRowContext(ctx, query, collection.Name, collection.Description, collection.OwnerId, collection.URL, collection.IsTemporary).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *CollectionRepo) UpdateCollection(ctx context.Context, collection *model.CollectionInput) error {
	query := `UPDATE public.collection 
              SET name = $1, description = $2, photo_url = $3, is_temporary = $4
              WHERE id = $5`

	result, err := r.db.ExecContext(ctx, query, collection.Name, collection.Description, collection.URL, collection.IsTemporary, collection.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (r *CollectionRepo) UpsertCollectionItem(ctx context.Context, tx *sql.Tx, collectionID, realEstateID int64) error {
	query := `INSERT INTO collection_item (collection_id, real_estate_id) VALUES ($1, $2) 
	ON CONFLICT (collection_id, real_estate_id) DO NOTHING`
	_, err := tx.ExecContext(ctx, query, collectionID, realEstateID)
	if err != nil {
		return err
	}

	return nil
}

func (r *CollectionRepo) DeleteCollection(ctx context.Context, tx *sql.Tx, collectionID int64) error {
	query := `UPDATE public.collection SET is_deleted = true WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, collectionID)
	if err != nil {
		return err
	}

	return nil
}

func (r *CollectionRepo) DeleteCollectionItem(ctx context.Context, tx *sql.Tx, collectionID, realEstateID int64) error {
	query := `DELETE FROM public.collection_item WHERE collection_id = $1 AND real_estate_id = $2`

	_, err := tx.ExecContext(ctx, query, collectionID, realEstateID)
	if err != nil {
		return err
	}

	return nil
}

func (r *CollectionRepo) GetCollection(ctx context.Context, filter *model.CollectionFilter) (*model.Collection, error) {
	query := `SELECT c.id, c.name, c.description, c.url, c.manager_id, c.created_at, c.photo_url, c.is_temporary FROM public.collection c WHERE c.id = $1 and c.is_deleted = false`
	var args []interface{}

	args = append(args, filter.Id)

	if filter.OwnerId != nil {
		query += " AND c.manager_id = $2"
		args = append(args, *filter.OwnerId)
	}

	if len(filter.Type) > 0 {
		statusPlaceholder := helper.GeneratePlaceholders(1, len(args)+1)
		query += fmt.Sprintf(` and
			EXISTS (
				SELECT 1 FROM (
					SELECT DISTINCT ON (ch.collection_id) ch.*
					FROM public.collection_history ch
					WHERE ch.collection_id = c.id
					ORDER BY ch.collection_id, ch.created_at DESC
				) latest
				WHERE latest.status_id = ANY(%s)
			)
		`, statusPlaceholder)
		args = append(args, pq.Array(filter.Type))
	}

	row := r.db.QueryRowContext(ctx, query, args...)

	var collection model.Collection
	err := row.Scan(&collection.ID, &collection.Name, &collection.Description, &collection.UUID, &collection.Manager.ID, &collection.CreatedAt, &collection.PhotoURL, &collection.IsTemporary)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &collection, nil
}

func (r *CollectionRepo) GetCollections(ctx context.Context, filter *model.CollectionFilter) (model.Collections, int, error) {
	var queryBuilder strings.Builder
	var params []interface{}
	var statusPlaceholder, ownerFilter, statusFilter string

	statement := `SELECT c.id, c.name, c.description, c.url, c.manager_id, c.created_at, c.photo_url, c.is_temporary `

	queryBuilder.WriteString(`
		FROM public.collection c
		WHERE is_deleted = false
	`)

	if len(filter.Type) > 0 {
		statusPlaceholder = helper.GeneratePlaceholders(1, len(params)+1)
		statusFilter = fmt.Sprintf(`
			EXISTS (
				SELECT 1 FROM (
					SELECT DISTINCT ON (ch.collection_id) ch.*
					FROM public.collection_history ch
					WHERE ch.collection_id = c.id
					ORDER BY ch.collection_id, ch.created_at DESC
				) latest
				WHERE latest.status_id = ANY(%s)
			)
		`, statusPlaceholder)
		params = append(params, pq.Array(filter.Type))
	}

	if filter.OwnerId != nil {
		if statusFilter != "" {
			ownerFilter = fmt.Sprintf("(%s OR c.manager_id = $%d)", statusFilter, len(params)+1)
			params = append(params, *filter.OwnerId)
		} else {
			ownerFilter = fmt.Sprintf("c.manager_id = $%d", len(params)+1)
			params = append(params, *filter.OwnerId)
		}
		queryBuilder.WriteString(" AND " + ownerFilter)
	} else if statusFilter != "" {
		queryBuilder.WriteString(" AND " + statusFilter)
	}

	if filter.Search != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND (c.name ILIKE $%d OR c.description ILIKE $%d)", len(params)+1, len(params)+2))
		params = append(params, "%"+*filter.Search+"%", "%"+*filter.Search+"%")
	}

	if filter.IsTemporary != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND (c.is_temporary = $%d)", len(params)+1))
		params = append(params, *filter.IsTemporary)
	}

	var totalCount int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) `+queryBuilder.String(), params...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	queryBuilder.WriteString(" ORDER BY c.created_at DESC")

	firstIndex := (filter.Page - 1) * filter.Rows
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Rows, firstIndex))

	rows, err := r.db.QueryContext(ctx, statement+queryBuilder.String(), params...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, model.ErrNotFound
		}

		return nil, 0, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var collections model.Collections
	for rows.Next() {
		var collection model.Collection
		err := rows.Scan(&collection.ID, &collection.Name, &collection.Description, &collection.UUID, &collection.Manager.ID, &collection.CreatedAt, &collection.PhotoURL, &collection.IsTemporary)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, 0, model.ErrNotFound
			}

			return nil, 0, err
		}
		collections = append(collections, collection)
	}

	return collections, totalCount, nil
}

func (r *CollectionRepo) SetStatus(ctx context.Context, tx *sql.Tx, collectionID, userId int64, statusID int) error {
	query := `INSERT INTO collection_history (collection_id, status_id, manager_id) 
	VALUES ($1, $2, $3)`
	_, err := tx.ExecContext(ctx, query, collectionID, statusID, userId)
	if err != nil {
		return err
	}

	return nil
}

func (r *CollectionRepo) DeleteCollectionItems(ctx context.Context, tx *sql.Tx, collectionID int64) error {
	query := `DELETE FROM public.collection_item WHERE collection_id = $1`
	_, err := tx.ExecContext(ctx, query, collectionID)
	return err
}

func (r *CollectionRepo) GetCollectionLastHistory(ctx context.Context, collectionId int64) (*model.CollectionHistory, error) {
	var history model.CollectionHistory

	query := `
		SELECT ch.id, ch.collection_id, ch.status_id, ch.created_at, ch.manager_id, 
		u.first_name, u.last_name, u.middle_name, u.phone, u.email, s.name
		FROM public.collection_history ch
		JOIN public."user" u ON u.id = ch.manager_id
		JOIN status s ON s.id = ch.status_id 
		WHERE ch.collection_id = $1
		ORDER BY ch.id DESC
		LIMIT 1;
	`

	err := r.db.QueryRowContext(ctx, query, collectionId).Scan(
		&history.ID,
		&history.CollectionId,
		&history.Status.ID,
		&history.CreatedAt,
		&history.Manager.ID,
		&history.Manager.FirstName,
		&history.Manager.LastName,
		&history.Manager.MiddleName,
		&history.Manager.Phone,
		&history.Manager.Email,
		&history.Status.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &history, nil
}

func (r *CollectionRepo) GetCollectionByUUID(ctx context.Context, uuid string) (*model.Collection, error) {
	query := `SELECT id, name, description, url, manager_id, created_at, photo_url, is_temporary FROM public.collection WHERE url = $1`
	row := r.db.QueryRowContext(ctx, query, uuid)

	var collection model.Collection
	err := row.Scan(&collection.ID, &collection.Name, &collection.Description, &collection.UUID, &collection.Manager.ID, &collection.CreatedAt, &collection.PhotoURL, &collection.IsTemporary)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &collection, nil
}

func (r *CollectionRepo) SetCollectionUUID(ctx context.Context, collectionId int64, uuid string) error {
	query := `UPDATE public.collection SET url = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, uuid, collectionId)
	return err
}

func (r *CollectionRepo) GetLatestHistoryByCollectionID(ctx context.Context, collectionID int64) (*model.CollectionHistory, error) {
	query := `
		SELECT ch.id, ch.status_id, ch.collection_id, ch.manager_id, ch.created_at
		FROM public.collection_history ch
		WHERE ch.collection_id = $1
		ORDER BY ch.id DESC
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, collectionID)

	var history model.CollectionHistory
	err := row.Scan(&history.ID, &history.Status.ID, &history.CollectionId, &history.Manager.ID, &history.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &history, nil
}

func (r *CollectionRepo) InsertTranslation(ctx context.Context, tx *sql.Tx, collectionID int64, langID int, name, description string) error {
	query := `
		INSERT INTO collection_translations (collection_id, lang_id, "name", description)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (collection_id, lang_id) DO UPDATE
		SET "name" = EXCLUDED.name,
		    description = EXCLUDED.description
	`
	_, err := tx.ExecContext(ctx, query, collectionID, langID, name, description)
	return err
}

func (r *CollectionRepo) GetCollectionTranslations(ctx context.Context, collectionID int64) ([]model.CollectionTranslation, error) {
	query := `
	SELECT ct.lang_id, l.code, ct.name, ct.description
	FROM collection_translations ct
	inner join "language" l on l.id = ct.lang_id 
	WHERE ct.collection_id = $1
	ORDER BY ct.lang_id
	`
	rows, err := r.db.QueryContext(ctx, query, collectionID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var translations []model.CollectionTranslation
	for rows.Next() {
		var t model.CollectionTranslation
		if err := rows.Scan(&t.Language.ID, &t.Language.Code, &t.Name, &t.Description); err != nil {
			return nil, err
		}
		translations = append(translations, t)
	}

	return translations, nil
}
