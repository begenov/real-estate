package model

import "time"

type PriceHistoryEntry struct {
	ID           int64     `json:"id"`
	RealEstateID int64     `json:"real_estate_id"`
	Price        float64   `json:"price"`
	CreatedAt    time.Time `json:"created_at"`
}

type PriceHistoryResponse struct {
	History []PriceHistoryEntry `json:"history"`
}
