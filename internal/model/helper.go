package model

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

func (r *RealEstateInput) ParseCompletionDate() (time.Time, error) {
	if r.CompletionDate == nil {
		return time.Time{}, nil
	}

	return time.Parse("2006-01-02", *r.CompletionDate)
}

func (r *RealEstateInput) Validate() error {
	if r.OwnerId == 0 {
		return ErrInvalidIdentifierNotBeEmpty
	}

	if r.Area <= 0 {
		return ErrInvalidArea
	}

	if r.Price <= 0 {
		return ErrInvalidPrice
	}

	if len(r.NewPhotos) == 0 && r.ID == 0 {
		return ErrInvalidPhotoURLs
	}

	if r.Purpose.ID == 0 {
		return ErrInvalidPurpose
	}

	if r.Purpose.ID != RealEstatePurpose_Sale && r.Purpose.ID != RealEstatePurpose_Rent {
		return ErrInvalidPurposeID
	}

	if r.Type.ID == 0 {
		return ErrInvalidType
	}

	if r.Type.ID != RealEstateType_Apartment && r.Type.ID != RealEstateType_Townhouses &&
		r.Type.ID != RealEstateType_Commercial && r.Type.ID != RealEstateType_Land &&
		r.Type.ID != RealEstateType_House && r.Type.ID != RealEstateType_SecondaryHouse &&
		r.Type.ID != RealEstateType_Villas {
		return ErrInvalidTypeID
	}

	if r.Latitude == 0 || r.Longitude == 0 {
		return ErrInvalidLatitudeLongitude
	}

	if r.Latitude < -90 || r.Latitude > 90 {
		return ErrInvalidLatitudeRange
	}
	if r.Longitude < -180 || r.Longitude > 180 {
		return ErrInvalidLongitudeRange
	}

	//_, err := time.Parse(time.DateOnly, r.CompletionDate)
	//if err != nil {
	//	return ErrInvalidCompletionDate
	//}

	if len(r.Translations) == 0 {
		return ErrTranslationsRequired
	}

	return nil
}

func (u *UserCreateInput) Validate() error {
	if utf8.RuneCountInString(strings.TrimSpace(u.Username)) == 0 {
		return ErrUsernameRequired
	}

	if utf8.RuneCountInString(strings.TrimSpace(u.Username)) < 3 {
		return ErrUsernameTooShort
	}

	if utf8.RuneCountInString(strings.TrimSpace(u.Email)) == 0 {
		return ErrEmailRequired
	}

	//emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	//re := regexp.MustCompile(emailRegex)
	//if !re.MatchString(u.Email) {
	//	return ErrInvalidEmailFormat
	//}

	if utf8.RuneCountInString(strings.TrimSpace(u.Phone)) == 0 {
		return ErrPhoneRequired
	}

	//TODO
	//if utf8.RuneCountInString(strings.TrimSpace(u.Password)) == 0 {
	//	return ErrPasswordRequired
	//}
	//
	//if utf8.RuneCountInString(strings.TrimSpace(u.Password)) < 8 {
	//	return ErrPasswordTooShort
	//}

	if utf8.RuneCountInString(strings.TrimSpace(u.FirstName)) == 0 {
		return ErrFirstNameRequired
	}

	if utf8.RuneCountInString(strings.TrimSpace(u.LastName)) == 0 {
		return ErrLastNameRequired
	}

	return nil
}

func (c *CollectionInput) Validate() error {

	if utf8.RuneCountInString(strings.TrimSpace(c.Name)) == 0 {
		return ErrNameIsRequired
	}

	if utf8.RuneCountInString(strings.TrimSpace(c.Description)) == 0 {
		return ErrDescriptionIsRequired
	}

	//if len(c.Items) == 0 {
	//	return errors.New("collection items is empty")
	//}

	if c.OwnerId == 0 {
		return ErrInvalidIdentifierNotBeEmpty
	}

	return nil
}

func (f *RealEstateFilter) Validate() error {
	if f.PriceMin != nil && f.PriceMax != nil && *f.PriceMin > *f.PriceMax {
		return ErrPriceMinGreaterThanMax
	}

	if f.Rooms != nil {
		for i := range f.Rooms {
			if len(f.Rooms[i]) == 0 {
				return ErrRoomsMustBePositive
			}
		}
	}

	if f.RegionID != nil && *f.RegionID <= 0 {
		return ErrInvalidRegionID
	}

	for i := range f.Type {
		if _, exists := validRealEstateTypes[f.Type[i]]; !exists {
			return ErrInvalidType
		}
	}

	for i := range f.Purpose {
		if _, exists := validRealEstatePurposes[f.Purpose[i]]; !exists {
			return ErrInvalidPurpose
		}
	}

	if f.Rows <= 0 {
		return ErrRowsMustBePositive
	}

	if f.Page <= 0 {
		return ErrPageMustBePositive
	}

	for _, status := range f.Status {
		if status <= 0 {
			return fmt.Errorf("%w: %d", ErrInvalidStatus, status)
		}
	}

	return nil
}

func (f *CollectionFilter) Validate() error {
	if f.Rows <= 0 {
		return ErrRowsMustBePositive
	}

	if f.Page <= 0 {
		return ErrPageMustBePositive
	}

	return nil
}
