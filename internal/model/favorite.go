package model

import "time"

type Favorite struct {
	RealEstateID int64       `json:"real_estate_id"`
	RealEstate   *RealEstate `json:"real_estate,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

type FavoriteInput struct {
	RealEstateID int64 `json:"real_estate_id" binding:"required"`
}

type FavoritesResponse struct {
	Favorites []Favorite `json:"favorites"`
	Total     int        `json:"total"`
}
