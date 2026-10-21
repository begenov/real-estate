package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/begenov/real-estate/internal/model"
)

type IFavoriteRepo interface {
	Add(ctx context.Context, userID, realEstateID int64) error
	Remove(ctx context.Context, userID, realEstateID int64) error
	GetByUserID(ctx context.Context, userID int64, page, rows int) ([]model.Favorite, int, error)
	Exists(ctx context.Context, userID, realEstateID int64) (bool, error)
}

type FavoriteRepo struct {
	db *sql.DB
}

func NewFavoriteRepo(db *sql.DB) IFavoriteRepo {
	return &FavoriteRepo{db: db}
}

func (r *FavoriteRepo) Add(ctx context.Context, userID, realEstateID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO public.user_favorites (user_id, real_estate_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, realEstateID)
	if err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}
	return nil
}

func (r *FavoriteRepo) Remove(ctx context.Context, userID, realEstateID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM public.user_favorites WHERE user_id = $1 AND real_estate_id = $2`,
		userID, realEstateID)
	if err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}
	return nil
}

func (r *FavoriteRepo) GetByUserID(ctx context.Context, userID int64, page, rows int) ([]model.Favorite, int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.user_favorites WHERE user_id = $1`, userID).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count favorites: %w", err)
	}

	offset := (page - 1) * rows
	dbRows, err := r.db.QueryContext(ctx,
		`SELECT real_estate_id, created_at FROM public.user_favorites
		 WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, rows, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get favorites: %w", err)
	}
	defer func() { _ = dbRows.Close() }()

	var favorites []model.Favorite
	for dbRows.Next() {
		var f model.Favorite
		if err := dbRows.Scan(&f.RealEstateID, &f.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan favorite: %w", err)
		}
		favorites = append(favorites, f)
	}

	return favorites, count, nil
}

func (r *FavoriteRepo) Exists(ctx context.Context, userID, realEstateID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM public.user_favorites WHERE user_id = $1 AND real_estate_id = $2)`,
		userID, realEstateID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("exists favorite: %w", err)
	}
	return exists, nil
}
