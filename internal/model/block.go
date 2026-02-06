package model

import "time"

type Block struct {
	ID            int64           `json:"id"`
	PageID        int64           `json:"page_id"`
	BlockContents []*BlockContent `json:"block_contents"`

	Translation map[string]BlockTranslation `json:"translation"`
	Type        BlockType                   `json:"type"`
}

type BlockTranslation struct {
	ID          int64    `json:"id"`
	BlockID     int64    `json:"block_id"`
	Language    Language `json:"language"`
	Name        string   `json:"name"`
	Description string   `json:"description"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type BlockContent struct {
	ID          int64                       `json:"id"`
	Translation map[string]BlockTranslation `json:"translation"`
	BLockID     int64                       `json:"block_id"`
	ImageURL    *string                     `json:"image_url"`
	CreatedAt   *time.Time                  `json:"created_at"`
	UpdatedAt   *time.Time                  `json:"updated_at"`
}

type BlockType struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
