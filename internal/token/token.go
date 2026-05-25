package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret     string
	expiration time.Duration
}

func NewManager(secret string) *Manager {
	return &Manager{
		secret:     secret,
		expiration: time.Hour * 24 * 30,
	}
}

func (m *Manager) Generate(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiration)),
	})

	signed, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", err
	}

	return signed, nil
}

func (m *Manager) Verify(tokenString string) (string, error) {

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(m.secret), nil
	})
	if err != nil {
		return "", err
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}

	return userId, nil
}
