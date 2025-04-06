package service

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Session struct {
	Token  string
	UserID int
	Role   string
}

func NewSession(userID int, role string) (*Session, error) {
	secret := os.Getenv("ADMIN_SECRET")
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": role,
		"iss": "media-organizer",
		"aud": userID,
		"exp": time.Now().Add(4 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	token, err := claims.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return &Session{
		Token:  token,
		UserID: userID,
		Role:   role,
	}, nil
}
