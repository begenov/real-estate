package model

import "errors"

var (
	//TODO general errors
	ErrNotFound                    = errors.New("not found")
	ErrInvalidIdentifier           = errors.New("invalid identifier")
	ErrInvalidIdentifierNotBeEmpty = errors.New("identifier cannot be empty")
	ErrAccessDenied                = errors.New("access denied")
	ErrDataIsEmpty                 = errors.New("data cannot be empty")
	ErrInvalidCompletionDate       = errors.New("invalid completion date format")
	ErrBadRequestQuery             = errors.New("invalid request query")
	ErrInvalidType                 = errors.New("type must be specified")
	ErrInvalidTypeID               = errors.New("invalid type ID")
	ErrNameIsRequired              = errors.New("name is required")
	ErrDescriptionIsRequired       = errors.New("description is required")
	ErrRowsMustBePositive          = errors.New("rows per page must be a positive number")
	ErrPageMustBePositive          = errors.New("page number must be a positive number")

	//TODO file errors
	ErrBucketNotExist   = errors.New("bucket does not exist")
	ErrFileNotSupported = errors.New("file not supported")
	ErrFileBadRequest   = errors.New("file bad request")

	//TODO real-estate errors
	ErrInvalidArea                = errors.New("area must be greater than zero")
	ErrInvalidPrice               = errors.New("price must be greater than zero")
	ErrInvalidLatitudeLongitude   = errors.New("latitude and longitude must be specified and valid")
	ErrInvalidPurpose             = errors.New("purpose must be specified")
	ErrInvalidPurposeID           = errors.New("invalid purpose ID")
	ErrInvalidLatitudeRange       = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitudeRange      = errors.New("longitude must be between -180 and 180")
	ErrInvalidPhotoURLs           = errors.New("at least one photo URL must be provided")
	ErrPriceMinGreaterThanMax     = errors.New("minimum price cannot be greater than maximum price")
	ErrRoomsMustBePositive        = errors.New("number of rooms must be a positive number")
	ErrInvalidRegionID            = errors.New("region_id must be a positive number")
	ErrInvalidStatus              = errors.New("invalid real estate status")
	ErrIdMustBeGreaterThanZero    = errors.New("id cannot be greater than zero")
	ErrTranslationsRequired       = errors.New("translations required")
	ErrRealEstateStatus           = errors.New("invalid real estate status")
	ErrRealEstateStatusNotCreated = errors.New("cannot update real estate: status is not created")

	//TODO user errors
	ErrUsernameRequired      = errors.New("username is required")
	ErrUsernameTooShort      = errors.New("username must be at least 3 characters long")
	ErrEmailRequired         = errors.New("email is required")
	ErrInvalidEmailFormat    = errors.New("invalid email format")
	ErrPhoneRequired         = errors.New("phone number is required")
	ErrPasswordRequired      = errors.New("password is required")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters long")
	ErrFirstNameRequired     = errors.New("first name is required")
	ErrLastNameRequired      = errors.New("last name is required")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrSignIn                = errors.New("invalid username or password")
	ErrUserNotActive         = errors.New("user is not active")

	//TODO token errors
	ErrTokenIsExpired     = errors.New("token is expired")
	ErrUserCtxNotFound    = errors.New("user ctx not found")
	ErrInvalidUserCtxType = errors.New("invalid user ctx type")

	//TODO collection errors
	ErrCollectionItemsIsEmpty       = errors.New("collection items is empty")
	ErrInvalidCollectionStatus      = errors.New("collection cannot be deleted due to its current status")
	ErrRealEstateCollection         = errors.New("failed to get real estate")
	ErrRealEstateStatusNotAvailable = errors.New("real estate status is not available")
	ErrRealEstateHistory            = errors.New("failed to get real estate history")
)

type APIError struct {
	Status  int
	Message string
	Err     error
	Details any
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) StatusCode() int {
	return e.Status
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func NewAPIError(status int, message string, err error) *APIError {
	return &APIError{
		Status:  status,
		Message: message,
		Err:     err,
	}
}

func NewAPIErrorWithDetails(status int, message string, err error, details any) *APIError {
	return &APIError{
		Status:  status,
		Message: message,
		Err:     err,
		Details: details,
	}
}
