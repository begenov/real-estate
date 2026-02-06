package model

import "time"

type Region struct {
	ID        int64      `json:"id"`
	NameRu    string     `json:"name_ru"`
	NameEn    string     `json:"name_en"`
	NameDe    string     `json:"name_de"`
	NameTr    string     `json:"name_tr"`
	Country   *Country   `json:"country"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type Country struct {
	ID          int64      `json:"id"`
	NameRu      string     `json:"name_ru"`
	NameEn      string     `json:"name_en"`
	NameDe      string     `json:"name_de"`
	NameTr      string     `json:"name_tr"`
	Code        string     `json:"code"`
	PhoneNumber string     `json:"phone_number"`
	FlagURL     *string    `json:"flag_url"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type District struct {
	ID        int64      `json:"id"`
	NameRu    string     `json:"name_ru"`
	NameEn    string     `json:"name_en"`
	NameDe    string     `json:"name_de"`
	NameTr    string     `json:"name_tr"`
	Region    *Region    `json:"region"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
