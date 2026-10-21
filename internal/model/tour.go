package model

import "time"

const (
	TourStatusNew        = 1
	TourStatusInProgress = 2
	TourStatusCompleted  = 3
	TourStatusCancelled  = 4
)

type TourRequest struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	DesiredDate  time.Time `json:"desired_date"`
	RealEstateID *int64    `json:"real_estate_id,omitempty"`
	StatusID     int       `json:"status_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TourRequestInput struct {
	Name         string    `json:"name" binding:"required"`
	Phone        string    `json:"phone" binding:"required"`
	DesiredDate  time.Time `json:"desired_date" binding:"required"`
	RealEstateID *int64    `json:"real_estate_id"`
}

type TourRequestFilter struct {
	Page     int
	Rows     int
	StatusID *int
}

type TourRequestsResponse struct {
	Tours []TourRequest `json:"tours"`
	Total int           `json:"total"`
}

type UpdateTourStatusInput struct {
	StatusID int `json:"status_id" binding:"required"`
}
