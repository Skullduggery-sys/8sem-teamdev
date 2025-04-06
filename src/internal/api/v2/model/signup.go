package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	svcModel "git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
)

type SignUpRequest struct {
	Name     string `json:"name"`
	Login    string `json:"login"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

func ParseSignUpRequest(r *http.Request, adminSecret string) (*svcModel.User, error) {
	var req SignUpRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("request body cannot be read: %w", err)
	}

	if err = json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("request cannot be unmarshalled: %w", err)
	}

	switch {
	case req.Name == "":
		return nil, fmt.Errorf("name is absent")
	case req.Login == "":
		return nil, fmt.Errorf("login is absent")
	case req.Role == "":
		return nil, fmt.Errorf("role is absent")
	case req.Password == "":
		return nil, fmt.Errorf("password is absent")
	}

	return &svcModel.User{
		Name:        req.Name,
		Login:       req.Login,
		Role:        req.Role,
		Password:    req.Password,
		AdminSecret: adminSecret,
	}, nil
}
