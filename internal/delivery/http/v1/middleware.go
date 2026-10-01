package v1

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	goRedis "github.com/redis/go-redis/v9"
)

const (
	authorizationHeader = "Authorization"

	userCtx = "userId"
)

const (
	publicCacheTTL     = 5 * time.Minute
	publicPageCacheTTL = 30 * time.Minute
)

type cacheEntry struct {
	Status      int             `json:"status"`
	ContentType string          `json:"content_type"`
	Body        json.RawMessage `json:"body"`
}

type cacheWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *cacheWriter) Write(b []byte) (int, error) {
	if w.body != nil {
		_, _ = w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (h *Handler) parseAuthHeader(c *gin.Context) (*model.TokenDetails, error) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return nil, errors.New("empty auth header")
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("invalid auth header")
	}

	if len(headerParts[1]) == 0 {
		return nil, errors.New("token is empty")
	}

	return h.tokenManager.ParseAccessToken(headerParts[1])
}

func (h *Handler) userIdentity(c *gin.Context) {
	tokenDetails, err := h.parseAuthHeader(c)
	if err != nil {
		_ = c.Error(model.NewAPIError(http.StatusUnauthorized, model.ErrTokenIsExpired.Error(), err))
		c.Abort()
		return
	}

	if tokenDetails.AtExpires < time.Now().UTC().Unix() {
		_ = c.Error(model.NewAPIError(http.StatusUnauthorized, model.ErrTokenIsExpired.Error(), model.ErrTokenIsExpired))
		c.Abort()
		return
	}

	if err := h.userService.VerifyAccessToken(c.Request.Context(), tokenDetails.UserId, tokenDetails.AccessUuid, tokenDetails.SessionUuid); err != nil {
		_ = c.Error(model.NewAPIError(http.StatusUnauthorized, model.ErrTokenIsExpired.Error(), err))
		c.Abort()
		return
	}

	c.Set(userCtx, tokenDetails.UserId)
}

func (h *Handler) cachePublicGet() gin.HandlerFunc {
	return h.cachePublicGetWithTTL(publicCacheTTL)
}

func (h *Handler) cachePublicGetWithTTL(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.cache == nil || c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		cacheKey := buildCacheKey(c)
		if ok := h.serveFromCache(c, cacheKey); ok {
			return
		}

		_, err, shared := h.cacheSF.Do(cacheKey, func() (interface{}, error) {
			writer := &cacheWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
			c.Writer = writer
			c.Next()

			if writer.Status() != http.StatusOK {
				return nil, nil
			}

			contentType := c.Writer.Header().Get("Content-Type")
			if contentType == "" {
				contentType = "application/json"
			}

			entry := cacheEntry{
				Status:      writer.Status(),
				ContentType: contentType,
				Body:        json.RawMessage(writer.body.Bytes()),
			}

			if err := h.cache.Save(c.Request.Context(), cacheKey, entry, ttl); err != nil {
				logger.Warn("cache save error: ", err)
			}
			return nil, nil
		})
		if err != nil {
			logger.Warn("cache singleflight error: ", err)
		}
		if shared {
			if ok := h.serveFromCache(c, cacheKey); ok {
				return
			}
			c.Next()
		}
	}
}

func (h *Handler) serveFromCache(c *gin.Context, cacheKey string) bool {
	cached, err := h.cache.Get(c.Request.Context(), cacheKey)
	if err == nil && cached != "" {
		var entry cacheEntry
		if err := json.Unmarshal([]byte(cached), &entry); err == nil {
			c.Data(entry.Status, entry.ContentType, entry.Body)
			c.Abort()
			return true
		}
	} else if err != nil && !errors.Is(err, goRedis.Nil) {
		logger.Warn("cache get error: ", err)
	}
	return false
}

func buildCacheKey(c *gin.Context) string {
	key := c.Request.Method + ":" + c.Request.URL.Path + "?" + c.Request.URL.RawQuery
	if lang := c.GetHeader("Accept-Language"); lang != "" {
		key += ":" + lang
	}

	hash := sha1.Sum([]byte(key))
	return "public_cache:" + hex.EncodeToString(hash[:])
}

func getUserId(c *gin.Context) (int64, error) {
	return getIdByContext(c, userCtx)
}

func getIdByContext(c *gin.Context, context string) (int64, error) {
	idFromCtx, ok := c.Get(context)
	if !ok {
		return 0, model.ErrUserCtxNotFound
	}

	id, ok := idFromCtx.(int64)
	if !ok {
		return 0, model.ErrInvalidUserCtxType
	}

	return id, nil
}

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

	model.ErrFileBadRequest:             http.StatusBadRequest,
	model.ErrFileNotSupported:           http.StatusBadRequest,
	model.ErrBucketNotExist:             http.StatusBadRequest,
	model.ErrRealEstateStatusNotCreated: http.StatusBadRequest,

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
