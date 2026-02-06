package model

import "time"

type Language struct {
	ID        int        `json:"id"`
	Code      string     `json:"code"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

var LanguageMap = map[string]int{
	"ru": 1,
	"en": 2,
	"de": 3,
	"tr": 4,
}
