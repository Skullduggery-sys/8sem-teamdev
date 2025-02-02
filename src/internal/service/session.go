package service

import (
	"github.com/google/uuid"
)

type Session struct {
	Token  string
	UserID int
	Role   string
}

func NewSession(userID int, role string) *Session {
	return &Session{
		Token:  uuid.NewString(),
		UserID: userID,
		Role:   role,
	}
}
