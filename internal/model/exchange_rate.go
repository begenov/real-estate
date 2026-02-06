package model

import "time"

type ExchangeRate struct {
	CurrencyCode string    `db:"currency_code" json:"currency_code"`
	Rate         float64   `db:"rate" json:"rate"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
