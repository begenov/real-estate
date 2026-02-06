package model

import "time"

type Users []*User

type User struct {
	ID         int64      `json:"id"`
	Username   *string    `json:"username,omitempty"`
	Email      *string    `json:"email,omitempty"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	MiddleName string     `json:"middle_name"`
	Phone      string     `json:"phone"`
	Roles      []Role     `json:"roles,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
	PhotoURL   *string    `json:"photo_url,omitempty"`
	IsActive   bool       `json:"is_active"`
	IsAdmin    *bool      `json:"is_admin,omitempty"`

	Password *string `json:"-"`
	OwnerId  *int64  `json:"-"`
}
