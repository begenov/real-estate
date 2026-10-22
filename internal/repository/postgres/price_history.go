package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/begenov/real-estate/internal/model"
)

type IPriceHistoryRepo interface {
	Add(ctx context.Context, tx *sql.Tx, realEstateID int64, price float64) error
	GetByRealEstateID(ctx context.Context, realEstateID int64) ([]model.PriceHistoryEntry, error)
	GetLatestPrice(ctx context.Context, realEstateID int64) (float64, bool, error)
}

type PriceHistoryRepo struct {
	db *sql.DB
}

func NewPriceHistoryRepo(db *sql.DB) IPriceHistoryRepo {
	return &PriceHistoryRepo{db: db}
}

func (r *PriceHistoryRepo) Add(ctx context.Context, tx *sql.Tx, realEstateID int64, price float64) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO public.real_estate_price_history (real_estate_id, price) VALUES ($1, $2)`,
		realEstateID, price)
	return err
}

func (r *PriceHistoryRepo) GetByRealEstateID(ctx context.Context, realEstateID int64) ([]model.PriceHistoryEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, real_estate_id, price, created_at
		 FROM public.real_estate_price_history
		 WHERE real_estate_id = $1
		 ORDER BY created_at ASC`,
		realEstateID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var entries []model.PriceHistoryEntry
	for rows.Next() {
		var e model.PriceHistoryEntry
		if err := rows.Scan(&e.ID, &e.RealEstateID, &e.Price, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (r *PriceHistoryRepo) GetLatestPrice(ctx context.Context, realEstateID int64) (float64, bool, error) {
	var price float64
	err := r.db.QueryRowContext(ctx,
		`SELECT price FROM public.real_estate_price_history
		 WHERE real_estate_id = $1
		 ORDER BY created_at DESC LIMIT 1`,
		realEstateID).Scan(&price)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return price, true, nil
}
