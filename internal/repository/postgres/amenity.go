package postgres

import (
	"context"
	"database/sql"
	"github.com/begenov/real-estate/internal/model"
)

type IAmenityRepo interface {
	Create(ctx context.Context, amenity *model.Amenity) error
	Update(ctx context.Context, amenity *model.Amenity) error
	Delete(ctx context.Context, amenityId int64) error
	GetAll(ctx context.Context) ([]*model.Amenity, error)
	GetByID(ctx context.Context, id int64) (*model.Amenity, error)
}

type AmenityRepo struct {
	db *sql.DB
}

func NewAmenityRepo(db *sql.DB) *AmenityRepo {
	return &AmenityRepo{db: db}
}

func (r *AmenityRepo) Create(ctx context.Context, amenity *model.Amenity) error {
	query := `
		INSERT INTO amenities (name_ru, name_en, name_de, name_tr, icon)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		amenity.NameRU, amenity.NameEN, amenity.NameDE, amenity.NameTR, amenity.Icon,
	).Scan(&amenity.ID, &amenity.CreatedAt)
}

func (r *AmenityRepo) Delete(ctx context.Context, amenityId int64) error {
	query := `
		UPDATE amenities
		SET is_deleted = TRUE
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, amenityId)
	return err
}

func (r *AmenityRepo) Update(ctx context.Context, amenity *model.Amenity) error {
	query := `
		UPDATE amenities
		SET name_ru = $1, name_en = $2, name_de = $3, name_tr = $4, icon = $5
		WHERE id = $6
	`
	_, err := r.db.ExecContext(ctx, query,
		amenity.NameRU, amenity.NameEN, amenity.NameDE, amenity.NameTR, amenity.Icon, amenity.ID,
	)
	return err
}

func (r *AmenityRepo) GetAll(ctx context.Context) ([]*model.Amenity, error) {
	query := `
		SELECT id, name_ru, name_en, name_de, name_tr, icon, created_at
		FROM amenities
		WHERE is_deleted = FALSE
		ORDER BY id
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var amenities []*model.Amenity
	for rows.Next() {
		var a model.Amenity
		if err := rows.Scan(
			&a.ID, &a.NameRU, &a.NameEN, &a.NameDE, &a.NameTR, &a.Icon, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		amenities = append(amenities, &a)
	}
	return amenities, nil
}

func (r *AmenityRepo) GetByID(ctx context.Context, id int64) (*model.Amenity, error) {
	query := `
		SELECT id, name_ru, name_en, name_de, name_tr, icon, created_at
		FROM amenities
		WHERE id = $1 AND is_deleted = FALSE
	`
	var a model.Amenity
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.NameRU, &a.NameEN, &a.NameDE, &a.NameTR, &a.Icon, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
