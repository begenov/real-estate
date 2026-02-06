package model

import "time"

type RealEstates []RealEstate

type RealEstate struct {
	ID             int64                   `json:"id"`
	Price          float64                 `json:"price"`
	Area           float64                 `json:"area"`
	Rooms          string                  `json:"rooms"`
	Region         Region                  `json:"region"`
	District       District                `json:"district"`
	Type           *Type                   `json:"type,omitempty"`
	Purpose        *RealEstatePurpose      `json:"purpose,omitempty"`
	Location       string                  `json:"location"` // Use the data type for geographic coordinates
	Latitude       float64                 `json:"latitude"`
	Longitude      float64                 `json:"longitude"`
	CompletionDate *time.Time              `json:"completion_date"`
	Manager        User                    `json:"creator"`
	LatestHistory  *RealEstateHistory      `json:"latest_history,omitempty"`
	Translations   []RealEstateTranslation `json:"translations"`
	Photos         []RealEstatePhoto       `json:"photos"`

	Address   string `json:"address"`
	Apartment string `json:"apartment"`

	Floor               *int       `json:"floor"`
	TotalFloors         *int       `json:"total_floors"`
	ParkingAvailable    *bool      `json:"parking_available"`
	BuiltYear           *int       `json:"built_year"`
	Amenities           []*Amenity `json:"amenities"`
	PricePerSquareMeter *float64   `json:"price_per_square_meter"`
	HasBalcony          *bool      `json:"has_balcony"`
	DistanceToSea       *float64   `json:"distance_to_sea"`

	PriceUSD float64 `json:"price_usd,omitempty"`
	PriceEUR float64 `json:"price_eur,omitempty"`

	SerialNumber string `json:"serial_number"`
}

type RealEstateHistory struct {
	ID           int       `json:"id"`
	RealEstateID int       `json:"real_estate_id"`
	Status       Status    `json:"status_id"`
	CreatedAt    time.Time `json:"created_at"`
	Manager      User      `json:"manager_id"`
}

type RealEstatePurpose struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type RealEstateTranslation struct {
	ID                 int      `json:"id"`
	RealEstateID       int      `json:"real_estate_id"`
	Description        string   `json:"description"`
	Name               string   `json:"name"`
	Language           Language `json:"language"`
	SummaryTitle       string   `json:"summary_title"`
	TermsAndConditions string   `json:"terms_and_conditions"`
}

type RealEstatePhoto struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}
