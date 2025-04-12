package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	svcModel "git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
)

type SignInRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token" example:"jwt-format token"`
}

func ParseSignInRequest(r *http.Request, adminSecret string) (*svcModel.User, error) {
	var req SignInRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("request body cannot be read: %w", err)
	}

	if err = json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("request cannot be unmarshalled: %w", err)
	}

	switch {
	case req.Login == "":
		return nil, fmt.Errorf("login is absent")
	case req.Password == "":
		return nil, fmt.Errorf("password is absent")
	}

	return &svcModel.User{
		Login:       req.Login,
		Password:    req.Password,
		AdminSecret: adminSecret,
	}, nil
}
