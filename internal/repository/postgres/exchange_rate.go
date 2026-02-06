package postgres

import (
	"context"
	"database/sql"
)

type IExchangeRateRepo interface {
	UpdateExchangeRates(ctx context.Context, rates map[string]float64) error
	GetExchangeRates(ctx context.Context) (map[string]float64, error)
}

type ExchangeRateRepo struct {
	db *sql.DB
}

func NewExchangeRateRepo(db *sql.DB) IExchangeRateRepo {
	return &ExchangeRateRepo{db: db}
}

func (r *ExchangeRateRepo) UpdateExchangeRates(ctx context.Context, rates map[string]float64) error {
	for code, rate := range rates {
		_, err := r.db.ExecContext(ctx, `
            INSERT INTO exchange_rates (currency_code, rate, updated_at)
            VALUES ($1, $2, NOW())
            ON CONFLICT (currency_code) DO UPDATE
            SET rate = EXCLUDED.rate, updated_at = NOW()
        `, code, rate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ExchangeRateRepo) GetExchangeRates(ctx context.Context) (map[string]float64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT currency_code, rate FROM exchange_rates`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	rates := make(map[string]float64)
	for rows.Next() {
		var code string
		var rate float64
		if err := rows.Scan(&code, &rate); err != nil {
			return nil, err
		}
		rates[code] = rate
	}
	return rates, nil
}
