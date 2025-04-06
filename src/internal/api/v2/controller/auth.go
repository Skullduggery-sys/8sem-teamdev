package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	reqModelPkg "git.iu7.bmstu.ru/vai20u117/testing/src/internal/api/v2/model"
	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	servicePkg "git.iu7.bmstu.ru/vai20u117/testing/src/internal/service"
	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenHeader = "X-User-Token"
)

type authService interface {
	GetUserTokenByAdmin(ctx context.Context, adminSecret, login string) (string, error)
	GetUserID(token string) (int, error)
	SignUp(ctx context.Context, user *model.User) (int, error)
	SignIn(ctx context.Context, user *model.User) (string, error)
	SignOut(ctx context.Context, token string) error
	Authorize(token string) error
}

type AuthHandler struct {
	service authService
}

func NewAuthHandler(service authService) *AuthHandler {
	return &AuthHandler{service: service}
}

// @Summary	Sign in
// @Description	sing in
// @Tags auth/v2
// @Param adminSecret query string true "Admin secret"
// @Param login query string true "User login"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /api/v2/sign-in [post]
func (h *AuthHandler) GetUserTokenByAdmin(ctx context.Context, adminSecret, login string) ([]byte, error) {
	token, err := h.service.GetUserTokenByAdmin(ctx, adminSecret, login)
	switch {
	case errors.Is(err, servicePkg.ErrAdminIsNotAuthtorized):
		slog.Warn("this user is not allowed to be an admin")
		return nil, fmt.Errorf("%w: %w", errInvalidArguments, err)
	case errors.Is(err, servicePkg.ErrNotFound):
		slog.Warn("user by login is not found", "user_login", login)
		return nil, fmt.Errorf("%w: %w", errNotFound, err)
	case err != nil:
		slog.Error("unexpected error occurred while signing in", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	tokenJSON, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		slog.Error("unexpected error occurred while marshaling token", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return tokenJSON, nil
}

func (h *AuthHandler) SignUp(ctx context.Context, user *model.User) ([]byte, error) {
	userID, err := h.service.SignUp(ctx, user)
	switch {
	case errors.Is(err, servicePkg.ErrAdminIsNotAuthtorized):
		slog.Warn("this user is not allowed to be an admin")
		return nil, errActionNotAuthorized
	case errors.Is(err, servicePkg.ErrLoginAlreadyExists):
		slog.Warn("user with such login already exists", "login", user.Login)
		return nil, fmt.Errorf("%w: %w", errInvalidArguments, err)
	case errors.Is(err, servicePkg.ErrGeneratingHash):
		slog.Warn("failed to generate password hash", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	case err != nil:
		slog.Error("unexpected error occurred while signing up", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	idJSON, err := json.Marshal(map[string]int{"id": userID})
	if err != nil {
		slog.Error("unexpected error occurred while marshaling id", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return idJSON, nil
}

func (h *AuthHandler) SignIn(ctx context.Context, user *model.User) ([]byte, error) {
	token, err := h.service.SignIn(ctx, user)
	switch {
	case errors.Is(err, servicePkg.ErrGeneratingHash):
		slog.Warn("failed to generate password hash", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	case errors.Is(err, servicePkg.ErrNotFound):
		slog.Warn("user by login is not found", "user_login", user.Login)
		return nil, fmt.Errorf("%w: %w", errNotFound, err)
	case errors.Is(err, servicePkg.ErrBadPassword):
		slog.Warn("user password is not matched", "user_login", user.Login)
		return nil, fmt.Errorf("%w: %w", errInvalidArguments, err)
	case errors.Is(err, servicePkg.ErrAdminIsNotAuthtorized):
		slog.Warn("this user is not allowed to be an admin")
		return nil, fmt.Errorf("%w: %w", errInvalidArguments, err)
	case err != nil:
		slog.Error("unexpected error occurred while signing in", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	tokenJSON, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		slog.Error("unexpected error occurred while marshaling token", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return tokenJSON, nil
}

func (h *AuthHandler) SignOut(ctx context.Context, token string) error {
	if err := h.service.SignOut(ctx, token); errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("token is not found")
		return errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while signing out", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

func (c *Controller) handleGetUserTokenRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		adminSecret := r.URL.Query().Get("admin_secret")
		login := r.URL.Query().Get("login")
		token, err := c.auth.GetUserTokenByAdmin(ctx, adminSecret, login)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(token); err != nil {
			writeError(w, fmt.Errorf("%w: writing token body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

// @Summary	Sign up
// @Description	sing up
// @Tags auth/v2
// @Param input body reqModelPkg.SignUpRequest true "User body"
// @Param admin_secret query string false "Admin auth secret"
// @Accept json
// @Success	201 {integer} int "ID"
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/sign-up [post]
func (c *Controller) handleSignUpRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		adminSecret := r.URL.Query().Get("admin_secret")
		user, err := reqModelPkg.ParseSignUpRequest(r, adminSecret)
		if err != nil {
			slog.Warn("failed to parse request", "error", err)
			writeError(w, err)
			break
		}

		id, err := c.auth.SignUp(ctx, user)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(id); err != nil {
			writeError(w, fmt.Errorf("%w: writing user_id body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusCreated)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

// @Summary	Sign in
// @Description	sing in
// @Tags auth/v2
// @Param input body reqModelPkg.SignInRequest true "User body"
// @Param admin_secret query string false "Admin auth secret"
// @Accept json
// @Success	200 {object} reqModelPkg.TokenResponse "Token"
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/sign-in [post]
func (c *Controller) handleSignInRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		adminSecret := r.URL.Query().Get("admin_secret")
		user, err := reqModelPkg.ParseSignInRequest(r, adminSecret)
		if err != nil {
			slog.Warn("failed to parse request", "error", err)
			writeError(w, err)
			break
		}

		token, err := c.auth.SignIn(ctx, user)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(token); err != nil {
			writeError(w, fmt.Errorf("%w: writing token body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

// @Summary	Sign out
// @Description	sing out
// @Tags auth/v2
// @Param X-User-Token header string true "JWT-format token"
// @Accept json
// @Success	200
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/sign-out [post]
func (c *Controller) handleSignOutRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		token := r.Header.Get(tokenHeader)
		if err := c.auth.SignOut(ctx, token); err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

func parseJWT(token string) (*servicePkg.Session, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ADMIN_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	return &servicePkg.Session{
		Token:  token,
		Role:   claims["sub"].(string),
		UserID: int(claims["aud"].(float64)),
	}, nil
}
