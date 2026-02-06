package v1

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
		logger.Error("h.parseAuthHeader(): ", err)
		newResponse(c, http.StatusUnauthorized, model.ErrTokenIsExpired.Error())
		return
	}

	if tokenDetails.AtExpires < time.Now().UTC().Unix() {
		newResponse(c, http.StatusUnauthorized, model.ErrTokenIsExpired.Error())
		return
	}

	err = h.userService.VerifyAccessToken(c.Request.Context(), tokenDetails.UserId, tokenDetails.AccessUuid, tokenDetails.SessionUuid)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, model.ErrTokenIsExpired.Error())
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
