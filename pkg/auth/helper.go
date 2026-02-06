package auth

import (
	"errors"
	"fmt"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
)

func (m *Manager) parseToken(tokenString, secretKey string, isAccessToken bool) (*model.TokenDetails, error) {
	token, err := m.parse(tokenString, secretKey)
	if err != nil {
		logger.Error("m.parseToken err:", err)
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		logger.Error("m.parseToken err:", err)
		return nil, errors.New("token is not valid")
	}

	var tokenDetails model.TokenDetails

	if err := m.claimToString(claims, "session_uuid", &tokenDetails.SessionUuid); err != nil {
		logger.Error("m.parseToken err:", err)
		return nil, err
	}
	if err := m.claimToInt(claims, "user_id", &tokenDetails.UserId); err != nil {
		logger.Error("m.parseToken err:", err)
		return nil, err
	}

	if isAccessToken {
		if err := m.claimToString(claims, "access_uuid", &tokenDetails.AccessUuid); err != nil {
			logger.Error("m.parseToken err:", err)
			return nil, err
		}

		if err := m.claimToInt64(claims, "exp", &tokenDetails.AtExpires); err != nil {
			logger.Error("m.parseToken err:", err)
			return nil, err
		}
	} else {
		if err := m.claimToString(claims, "refresh_uuid", &tokenDetails.RefreshUuid); err != nil {
			logger.Error("m.parseToken err:", err)
			return nil, err
		}

		if err := m.claimToInt64(claims, "exp", &tokenDetails.RtExpires); err != nil {
			logger.Error("m.parseToken err:", err)
			return nil, err
		}
	}

	return &tokenDetails, nil
}

func (m *Manager) claimToString(claims jwt.MapClaims, key string, dest *string) error {
	val, ok := claims[key].(string)

	if !ok {
		return fmt.Errorf("can't claim %s", key)
	}
	*dest = val
	return nil
}

func (m *Manager) claimToInt(claims jwt.MapClaims, key string, dest *int64) error {
	val, ok := claims[key].(string)
	if !ok {
		return fmt.Errorf("can't claim %s", key)
	}
	parsedVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf("can't parse %s", key)
	}
	*dest = parsedVal
	return nil
}

func (m *Manager) claimToInt64(claims jwt.MapClaims, key string, dest *int64) error {
	val, ok := claims[key].(string)
	if !ok {
		return fmt.Errorf("can't claim %s", key)
	}

	parsedVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf("can't parse %s", key)
	}

	*dest = parsedVal
	return nil
}

func (m *Manager) parse(token, secretKey string) (*jwt.Token, error) {
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return jwtToken, nil
}
