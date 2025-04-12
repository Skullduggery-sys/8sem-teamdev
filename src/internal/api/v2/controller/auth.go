package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	servicePkg "git.iu7.bmstu.ru/vai20u117/testing/src/internal/service"
)

const (
	tokenHeader = "X-User-Token"
)

type authService interface {
	GetUserByTGID(ctx context.Context, tgID string) (*model.User, error)
	SignUp(ctx context.Context, tgID string) (int, error)
}

type AuthHandler struct {
	service authService
}

func NewAuthHandler(service authService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) SignUp(ctx context.Context, tgID string) ([]byte, error) {
	if tgID == "" {
		slog.Warn("got empty token to sign-up")
		return nil, fmt.Errorf("%w: tg_id is empty", errInvalidArguments)
	}

	userID, err := h.service.SignUp(ctx, tgID)
	switch {
	case errors.Is(err, servicePkg.ErrLoginAlreadyExists):
		slog.Warn("such user already exists", "tg_id", tgID)
		return nil, errUserAlreadyExists
	case err != nil:
		slog.Error("unexpected error occurred while signing up", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	slog.Debug("registered new user", "tg_id", tgID, "id", userID)
	idJSON, err := json.Marshal(map[string]int{"id": userID})
	if err != nil {
		slog.Error("unexpected error occurred while marshaling id", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return idJSON, nil
}

// @Summary	Sign up
// @Description	sing up
// @Tags auth/v2
// @Param input body reqModelPkg.SignUpRequest true "User body"
// @Param admin_secret query string false "Admin auth secret"
// @Accept json
// @Success	201 {integer} int "ID"
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	409	{object} reqModelPkg.ErrorResponse "user already exists"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/sign-up [post]
func (c *Controller) handleSignUpRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		token := r.Header.Get(tokenHeader)

		userID, err := c.auth.SignUp(ctx, token)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(userID); err != nil {
			writeError(w, fmt.Errorf("%w: writing user_id body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusCreated)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}
