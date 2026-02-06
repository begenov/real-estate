package model

import "time"

type Collections []Collection

type Collection struct {
	ID            int64              `json:"id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	UUID          *string            `json:"uuid,omitempty"`
	Manager       User               `json:"manager"`
	RealEstates   RealEstates        `json:"real_estates"`
	LatestHistory *CollectionHistory `json:"latest_history,omitempty"`
	PhotoURL      string             `json:"photo_url"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	CollectionTranslation []CollectionTranslation `json:"translations"`

	IsTemporary bool `json:"is_temporary"`
}

type CollectionHistory struct {
	ID           int       `json:"id"`
	CollectionId int       `json:"collection_id"`
	Status       Status    `json:"status_id"`
	CreatedAt    time.Time `json:"created_at"`
	Manager      User      `json:"manager_id"`
}

type CollectionTranslation struct {
	Language    Language `json:"lang"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
}
