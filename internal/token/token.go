package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret     []byte
	expiration time.Duration
}

func NewManager(secret string) *Manager {
	return &Manager{
		secret:     []byte(secret),
		expiration: time.Hour * 24 * 30,
	}
}

func (m *Manager) Generate(userID string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiration)),
	})
	return t.SignedString(m.secret)
}

func (m *Manager) Verify(tokenString string) (string, error) {
	t, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		// Pin to HS256 exactly; the HMAC family alone is too wide.
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return "", err
	}
	return t.Claims.GetSubject()
}
