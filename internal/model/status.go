package model

const (
	Status_Created = iota + 1
	Status_Available
	Status_Sold
	Status_Archived
)

const (
	RealEstateType_Apartment = iota + 1
	RealEstateType_Commercial
	RealEstateType_Land
	RealEstateType_House
	RealEstateType_SecondaryHouse
	RealEstateType_Townhouses
	RealEstateType_Villas
)

const (
	RealEstatePurpose_Sale = iota + 1
	RealEstatePurpose_Rent
)

var validRealEstateTypes = map[int]struct{}{
	RealEstateType_Apartment:      {},
	RealEstateType_Commercial:     {},
	RealEstateType_Land:           {},
	RealEstateType_House:          {},
	RealEstateType_SecondaryHouse: {},
	RealEstateType_Townhouses:     {},
	RealEstateType_Villas:         {},
}

var validRealEstatePurposes = map[int]struct{}{
	RealEstatePurpose_Sale: {},
	RealEstatePurpose_Rent: {},
}

type Status struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Type struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
