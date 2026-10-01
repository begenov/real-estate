package auth

import (
	"errors"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenManager interface {
	ParseAccessToken(accessToken string) (*model.TokenDetails, error)
	ParseRefreshToken(refreshToken string) (*model.TokenDetails, error)
	NewTokenDetails(userId int64) (*model.TokenDetails, error)
}

type Manager struct {
	accessSignInKey  string
	refreshSignInKey string

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewManager(accessSignInKey, refreshSignInKey string, accessTokenTTL,
	refreshTokenTTL time.Duration) (*Manager, error) {
	if utf8.RuneCountInString(accessSignInKey) == 0 || utf8.RuneCountInString(refreshSignInKey) == 0 {
		return nil, errors.New("empty signing key")
	}
	return &Manager{
		accessSignInKey:  accessSignInKey,
		refreshSignInKey: refreshSignInKey,
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
	}, nil
}

func (m *Manager) ParseAccessToken(accessToken string) (*model.TokenDetails, error) {
	return m.parseToken(accessToken, m.accessSignInKey, true)
}

func (m *Manager) ParseRefreshToken(refreshToken string) (*model.TokenDetails, error) {
	return m.parseToken(refreshToken, m.refreshSignInKey, false)
}

func (m *Manager) NewTokenDetails(userId int64) (*model.TokenDetails, error) {
	tokenDetails := &model.TokenDetails{}

	tokenDetails.AccessUuid = uuid.New().String()
	tokenDetails.AtExpires = time.Now().Add(m.accessTokenTTL).Unix()

	tokenDetails.RefreshUuid = uuid.New().String()
	tokenDetails.RtExpires = time.Now().Add(m.refreshTokenTTL).Unix()

	tokenDetails.SessionUuid = uuid.New().String()

	atClaims := jwt.MapClaims{
		"access_uuid":  tokenDetails.AccessUuid,
		"session_uuid": tokenDetails.SessionUuid,
		"exp":          strconv.FormatInt(tokenDetails.AtExpires, 10),
		"user_id":      strconv.FormatInt(userId, 10),
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString([]byte(m.accessSignInKey))
	if err != nil {
		logger.Error("at.SignedString(): ", err)
		return nil, err
	}

	tokenDetails.AccessToken = accessToken

	rtClaims := jwt.MapClaims{
		"refresh_uuid": tokenDetails.RefreshUuid,
		"session_uuid": tokenDetails.SessionUuid,
		"exp":          strconv.FormatInt(tokenDetails.RtExpires, 10),
		"user_id":      strconv.FormatInt(userId, 10),
	}

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString([]byte(m.refreshSignInKey))
	if err != nil {
		logger.Error("rt.SignedString(): ", err)
		return nil, err
	}

	tokenDetails.RefreshToken = refreshToken

	return tokenDetails, nil
}
