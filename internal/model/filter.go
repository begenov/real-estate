package model

type UsersFilter struct {
	Page int
	Rows int

	UserId *int64
	RoleId *int
}

type UserFilter struct {
	Username *string
	Email    *string
	Id       *int64
}

type RealEstateFilter struct {
	PriceMin   *float64
	PriceMax   *float64
	Rooms      []string
	RegionID   *int
	Type       []int
	Purpose    []int
	SortBy     []string
	SortOrder  []string
	Page       int
	Rows       int
	DistrictID *int

	//
	Status   []int
	IsActive bool

	CollectionId *int64
	ID           *int64
	StatusID     *int64
}

type CollectionFilter struct {
	Search      *string
	Page        int
	Rows        int
	Id          *int64
	OwnerId     *int64
	Type        []int
	IsTemporary *bool
}

type RegionFilter struct {
	CountryId *int    `form:"country_id"`
	Name      *string `form:"name"`
	Page      int     `form:"page,default=1"`
	Rows      int     `form:"rows,default=10"`
}

type DistrictFilter struct {
	RegionId *int    `form:"region_id"`
	Name     *string `form:"name"`
	Page     int     `form:"page,default=1"`
	Rows     int     `form:"rows,default=10"`
}
