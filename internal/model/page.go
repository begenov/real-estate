package model

import (
	"time"
)

type Page struct {
	ID          int64                       `json:"id"`
	Translation map[string]*PageTranslation `json:"translation"`
	Slug        string                      `json:"slug"`

	Order     int64      `json:"order"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type PageTranslation struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Language    Language `json:"language"`

	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type PageHistory struct {
	ID     int64      `json:"id"`
	Status PageStatus `json:"status"`
	User   User       `json:"user"`

	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type PageType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PageStatus struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`

	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
