package http

import (
	"encoding/json"
	"errors"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if origin == "" {
			continue
		}
		allowed[origin] = struct{}{}
	}
	denyAll := len(allowed) == 0

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if denyAll {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			if _, ok := allowed[origin]; !ok {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Content-Type", "application/json")

		if c.Request.Method != "OPTIONS" {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusOK)
		}
	}
}

func errorHandlerMiddleware(c *gin.Context) {
	c.Next()

	if len(c.Errors) > 0 {
		err := c.Errors.Last().Err

		type httpError interface {
			StatusCode() int
		}
		if e, ok := err.(httpError); ok {
			writeAPIError(c, model.NewAPIError(e.StatusCode(), err.Error(), err))
			return
		}

		var apiErr *model.APIError
		if errors.As(err, &apiErr) {
			writeAPIError(c, apiErr)
			return
		}

		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			writeAPIError(c, model.NewAPIErrorWithDetails(http.StatusBadRequest, "validation failed", err, validationErrors.Error()))
			return
		}

		var syntaxErr *json.SyntaxError
		var unmarshalErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr), errors.As(err, &unmarshalErr):
			writeAPIError(c, model.NewAPIError(http.StatusBadRequest, "invalid json", err))
			return
		case errors.Is(err, io.EOF):
			writeAPIError(c, model.NewAPIError(http.StatusBadRequest, "empty request body", err))
			return
		}

		if status, ok := statusForError(err); ok {
			writeAPIError(c, model.NewAPIError(status, err.Error(), err))
			return
		}

		writeAPIError(c, model.NewAPIError(http.StatusInternalServerError, "Internal server error", err))
	}
}

func writeAPIError(c *gin.Context, err *model.APIError) {
	if err.Err != nil {
		logger.Error(err.Err)
	}
	if err.Details != nil {
		c.JSON(err.Status, gin.H{"error": err.Message, "details": err.Details})
		return
	}
	c.JSON(err.Status, gin.H{"error": err.Message})
}

var errorStatusMap = map[error]int{
	// general errors
	model.ErrNotFound: http.StatusNotFound,

	model.ErrInvalidIdentifier:           http.StatusBadRequest,
	model.ErrInvalidIdentifierNotBeEmpty: http.StatusBadRequest,
	model.ErrDataIsEmpty:                 http.StatusBadRequest,
	model.ErrInvalidCompletionDate:       http.StatusBadRequest,
	model.ErrBadRequestQuery:             http.StatusBadRequest,
	model.ErrInvalidType:                 http.StatusBadRequest,
	model.ErrInvalidTypeID:               http.StatusBadRequest,
	model.ErrNameIsRequired:              http.StatusBadRequest,
	model.ErrDescriptionIsRequired:       http.StatusBadRequest,
	model.ErrRowsMustBePositive:          http.StatusBadRequest,
	model.ErrPageMustBePositive:          http.StatusBadRequest,

	model.ErrAccessDenied: http.StatusForbidden,

	// user errors
	model.ErrUsernameRequired:      http.StatusBadRequest,
	model.ErrUsernameTooShort:      http.StatusBadRequest,
	model.ErrEmailRequired:         http.StatusBadRequest,
	model.ErrInvalidEmailFormat:    http.StatusBadRequest,
	model.ErrPhoneRequired:         http.StatusBadRequest,
	model.ErrPasswordRequired:      http.StatusBadRequest,
	model.ErrPasswordTooShort:      http.StatusBadRequest,
	model.ErrFirstNameRequired:     http.StatusBadRequest,
	model.ErrLastNameRequired:      http.StatusBadRequest,
	model.ErrUsernameAlreadyExists: http.StatusBadRequest,
	model.ErrEmailAlreadyExists:    http.StatusBadRequest,
	model.ErrSignIn:                http.StatusBadRequest,
	model.ErrUserNotActive:         http.StatusBadRequest,

	// file errors
	model.ErrFileBadRequest:             http.StatusBadRequest,
	model.ErrFileNotSupported:           http.StatusBadRequest,
	model.ErrBucketNotExist:             http.StatusBadRequest,
	model.ErrRealEstateStatusNotCreated: http.StatusBadRequest,

	// real estate errors
	model.ErrInvalidArea:              http.StatusBadRequest,
	model.ErrInvalidPrice:             http.StatusBadRequest,
	model.ErrInvalidLatitudeLongitude: http.StatusBadRequest,
	model.ErrInvalidPurpose:           http.StatusBadRequest,
	model.ErrInvalidPurposeID:         http.StatusBadRequest,
	model.ErrInvalidLatitudeRange:     http.StatusBadRequest,
	model.ErrInvalidLongitudeRange:    http.StatusBadRequest,
	model.ErrInvalidPhotoURLs:         http.StatusBadRequest,
	model.ErrPriceMinGreaterThanMax:   http.StatusBadRequest,
	model.ErrRoomsMustBePositive:      http.StatusBadRequest,
	model.ErrInvalidRegionID:          http.StatusBadRequest,
	model.ErrInvalidStatus:            http.StatusBadRequest,
	model.ErrIdMustBeGreaterThanZero:  http.StatusBadRequest,
	model.ErrTranslationsRequired:     http.StatusBadRequest,
	model.ErrRealEstateStatus:         http.StatusBadRequest,

	// collection errors
	model.ErrCollectionItemsIsEmpty:       http.StatusBadRequest,
	model.ErrInvalidCollectionStatus:      http.StatusBadRequest,
	model.ErrRealEstateCollection:         http.StatusBadRequest,
	model.ErrRealEstateStatusNotAvailable: http.StatusBadRequest,
	model.ErrRealEstateHistory:            http.StatusBadRequest,
}

func statusForError(err error) (int, bool) {
	for target, status := range errorStatusMap {
		if errors.Is(err, target) {
			return status, true
		}
	}
	return 0, false
}
