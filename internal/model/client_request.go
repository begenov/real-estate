package model

type UserSignInInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshInput struct {
	Token string `json:"token"`
}

type UserCreateInput struct {
	ID         int64   `json:"id"`
	Username   string  `json:"username"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
	Password   *string `json:"password"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	MiddleName string  `json:"middle_name"`
	PhotoURL   *string `json:"-"`
	OwnerId    *int64  `json:"-"`
	IsActive   bool    `json:"is_active"`
}

type RealEstateInput struct {
	ID             int64                        `json:"id"`
	Price          float64                      `json:"price"`
	Area           float64                      `json:"area"`
	Rooms          string                       `json:"rooms"`
	RegionID       int                          `json:"region_id"`
	DistrictID     int64                        `json:"district_id"`
	Type           Type                         `json:"type"`
	Purpose        RealEstatePurpose            `json:"purpose"`
	Latitude       float64                      `json:"latitude"`
	Longitude      float64                      `json:"longitude"`
	CompletionDate *string                      `json:"completion_date"`
	NewPhotos      []RealEstatePhoto            `json:"new_photos"`
	Translations   []RealEstateTranslationInput `json:"translations"`
	DeletedPhotos  []int64                      `json:"deleted_photos"`
	Address        string                       `json:"address"`
	Apartment      string                       `json:"apartment"`

	OwnerId  int64  `json:"-"`
	Location string `json:"-"`

	Floor               int      `json:"floor"`
	TotalFloors         int      `json:"total_floors"`
	ParkingAvailable    bool     `json:"parking_available"`
	BuiltYear           *int     `json:"built_year"`
	AmenityIDs          []int64  `json:"amenity_ids"`
	PricePerSquareMeter *float64 `json:"price_per_square_meter"`
	HasBalcony          bool     `json:"has_balcony"`
	DistanceToSea       float64  `json:"distance_to_sea"`
}

type RealEstateTranslationInput struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Language           Language `json:"language"`
	SummaryTitle       string   `json:"summary_title"`
	TermsAndConditions string   `json:"terms_and_conditions"`
}

type RealEstateResponse struct {
	RealEstate RealEstates `json:"real_estates"`
	Total      int         `json:"total"`
}

type CollectionInput struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Items        []int64 `json:"real_estates"`
	DeletedItems []int64 `json:"deleted_items"`
	OwnerId      int64   `json:"-"`
	URL          string  `json:"url"`
	IsTemporary  bool    `json:"is_temporary"`
}

type CollectionResponse struct {
	Collection *Collection `json:"collection,omitempty"`
	Total      int         `json:"total"`
}

type CollectionsResponse struct {
	Collections Collections `json:"collections,omitempty"`
	Total       int         `json:"total"`
}

type CollectionUUIDResponse struct {
	UUID string `json:"uuid"`
}

type UsersResponse struct {
	Users Users `json:"users"`
	Total int   `json:"total"`
}

type UpdateStatusRequest struct {
	StatusID int `json:"status_id" binding:"required"`
}
