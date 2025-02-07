package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/mail"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	servicePkg "git.iu7.bmstu.ru/vai20u117/testing/src/internal/service"
)

type authService interface {
	GetUserTokenByAdmin(ctx context.Context, adminSecret, login string) (string, error)
	GetUserID(token string) (int, error)
	SignUp(ctx context.Context, user *model.User) (int, error)
	SignIn(ctx context.Context, user *model.User) (string, error)
	SignOut(ctx context.Context, token string) error
	ResetPassword(ctx context.Context, login, email, oldPassword, newPassword string) error
	Handle2FA(ctx context.Context, email, code string) (*servicePkg.VeryfyUserMetadata, error)
	ClearVerifyEmail(email string)
}

type AuthHandler struct {
	service authService
}

func NewAuthHandler(service authService) *AuthHandler {
	return &AuthHandler{service: service}
}

// @Summary	Sign in
// @Description	sing in
// @Tags auth
// @Param adminSecret query string true "Admin secret"
// @Param login query string true "User login"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Success	201 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /sign-in [post]
func (h *AuthHandler) GetUserTokenByAdmin(ctx context.Context, adminSecret, login string) ([]byte, error) {
	token, err := h.service.GetUserTokenByAdmin(ctx, adminSecret, login)
	switch {
	case errors.Is(err, servicePkg.ErrAdminIsNotAuthtorized):
		slog.Warn("this user is not allowed to be an admin")
		return nil, errInvalidArguments
	case errors.Is(err, servicePkg.ErrNotFound):
		slog.Warn("user by login is not found", "user_login", login)
		return nil, errNotFound
	case err != nil:
		slog.Error("unexpected error occurred while signing in", "error", err)
		return nil, errInternal
	}

	tokenJSON, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		slog.Error("unexpected error occurred while marshaling token", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return tokenJSON, nil
}

// @Summary	Sign up
// @Description	sing up
// @Tags auth
// @Param input body model.User true "User body"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Success	201 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /sign-up [post]
func (h *AuthHandler) SignUp(ctx context.Context, user *model.User) ([]byte, error) {
	userID, err := h.service.SignUp(ctx, user)
	switch {
	case errors.Is(err, servicePkg.ErrAdminIsNotAuthtorized):
		slog.Warn("this user is not allowed to be an admin")
		return nil, errInvalidArguments
	case errors.Is(err, servicePkg.ErrLoginAlreadyExists):
		slog.Warn("user with such login already exists", "login", user.Login)
		return nil, errInvalidArguments
	case errors.Is(err, servicePkg.ErrGeneratingHash):
		slog.Warn("failed to generate password hash", "error", err)
		return nil, errInternal
	case errors.Is(err, servicePkg.ErrWaiting2FA):
		slog.Info("sign in wainting 2fa", "user_login", user.Login)
		return nil, errWaiting2FA
	case err != nil:
		slog.Error("unexpected error occurred while signing up", "error", err)
		return nil, errInternal
	}

	idJSON, err := json.Marshal(map[string]int{"id": userID})
	if err != nil {
		slog.Error("unexpected error occurred while marshaling id", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return idJSON, nil
}

// @Summary	Sign in
// @Description	sing in
// @Tags auth
// @Param input body model.User true "User body"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Success	201 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /sign-in [post]
func (h *AuthHandler) SignIn(ctx context.Context, user *model.User) ([]byte, error) {
	token, err := h.service.SignIn(ctx, user)
	switch {
	case errors.Is(err, servicePkg.ErrGeneratingHash):
		slog.Warn("failed to generate password hash", "error", err)
		return nil, errInternal
	case errors.Is(err, servicePkg.ErrNotFound):
		slog.Warn("user by login is not found", "user_login", user.Login)
		return nil, errNotFound
	case errors.Is(err, servicePkg.ErrBadPassword):
		slog.Warn("user password is not matched", "user_login", user.Login)
		return nil, errInvalidArguments
	case errors.Is(err, servicePkg.ErrAdminIsNotAuthtorized):
		slog.Warn("this user is not allowed to be an admin")
		return nil, errInvalidArguments
	case errors.Is(err, servicePkg.ErrWaiting2FA):
		slog.Info("sign in wainting 2fa", "user_login", user.Login)
		return nil, errWaiting2FA
	case err != nil:
		slog.Error("unexpected error occurred while signing in", "error", err)
		return nil, errInternal
	}

	tokenJSON, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		slog.Error("unexpected error occurred while marshaling token", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return tokenJSON, nil
}

// @Summary	Sign out
// @Description	sing out
// @Tags auth
// @Param token query string true "User token"
// @Accept json
// @Success	200
// @Failure	400	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /sign-out [post]
func (h *AuthHandler) SignOut(ctx context.Context, token string) error {
	if err := h.service.SignOut(ctx, token); errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("token is not found")
		return errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while signing out", "error", err)
		return errInternal
	}

	return nil
}

// @Summary	Verify2FA
// @Description	verify 2fa
// @Tags auth
// @Param email query string true "User email"
// @Param code query string true "Verification code"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /verify-2fa [post]
func (h *AuthHandler) Verify2FA(ctx context.Context, email, code string) ([]byte, error) {
	meta, err := h.service.Handle2FA(ctx, email, code)
	switch {
	case errors.Is(err, servicePkg.ErrGeneratingHash):
		slog.Warn("failed to generate password hash", "error", err)
		return nil, errInternal
	case errors.Is(err, servicePkg.ErrNotFound):
		slog.Warn("user by login is not found", "user_email", email)
		return nil, errNotFound
	case errors.Is(err, servicePkg.ErrBadPassword):
		slog.Warn("user password is not matched", "user_email", email)
		return nil, errInvalidArguments
	case errors.Is(err, servicePkg.ErrBadVeryfyCode):
		slog.Warn("verifcation code is not matched", "user_email", email)
		return nil, errInvalidArguments
	case errors.Is(err, servicePkg.ErrUserMetaNotFound):
		slog.Warn("user metadata is not found", "user_email", email)
		return nil, errInternal
	case errors.Is(err, servicePkg.ErrVerifyActionNotSaved):
		slog.Warn("user verify action is not found", "user_email", email)
		return nil, errInternal
	case errors.Is(err, servicePkg.ErrAdminIsNotAuthtorized):
		slog.Warn("this user is not allowed to be an admin")
		return nil, errInvalidArguments
	case err != nil:
		slog.Error("unexpected error occurred while verifying 2fa", "error", err)
		return nil, errInternal
	}

	var jsonData []byte
	var jsonErr error
	switch meta.Action {
	case servicePkg.UserMethodSignUp:
		jsonData, jsonErr = json.Marshal(map[string]int{"id": meta.User.ID})
	case servicePkg.UserMethodSignIn:
		jsonData, jsonErr = json.Marshal(map[string]string{"token": meta.Token})
	case servicePkg.UserMethodResetPassword:
		jsonData, jsonErr = json.Marshal("successfully reset password")
	default:
		return nil, errInternal
	}

	if jsonErr != nil {
		slog.Error("unexpected error occurred while marshaling data", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	h.service.ClearVerifyEmail(email)

	return jsonData, nil
}

// @Summary	Reset password
// @Description	reset password
// @Tags auth
// @Param login query string true "User login"
// @Param email query string true "User email"
// @Param oldPassword query string true "Old user password"
// @Param newPassword query string true "New user password"
// @Accept json
// @Success	200
// @Success	201	{object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /reset-password [post]
func (h *AuthHandler) ResetPassword(ctx context.Context, login, email, oldPassword, newPassword string) error {
	err := h.service.ResetPassword(ctx, login, email, oldPassword, newPassword)
	switch {
	case errors.Is(err, servicePkg.ErrGeneratingHash):
		slog.Warn("failed to generate password hash", "error", err)
		return errInternal
	case errors.Is(err, servicePkg.ErrNotFound):
		slog.Warn("user by login is not found", "user_login", login)
		return errNotFound
	case errors.Is(err, servicePkg.ErrBadPassword):
		slog.Warn("user password is not matched", "user_login", login)
		return errInvalidArguments
	case errors.Is(err, servicePkg.ErrAdminIsNotAuthtorized):
		slog.Warn("this user is not allowed to be an admin")
		return errInvalidArguments
	case errors.Is(err, servicePkg.ErrWaiting2FA):
		slog.Info("sign in wainting 2fa", "user_login", login)
		return errWaiting2FA
	case err != nil:
		slog.Error("unexpected error occurred while resetting password", "error", err)
		return errInternal
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

func (c *Controller) handleSignUpRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		user, err := parseUser(r)
		if err != nil {
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

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

func (c *Controller) handleSignInRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		user, err := parseUser(r)
		if err != nil {
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

func (c *Controller) handleSignOutRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		token := r.URL.Query().Get("token")
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

func (c *Controller) handleVerify2FARequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		email := r.URL.Query().Get("email")
		code := r.URL.Query().Get("code")
		data, err := c.auth.Verify2FA(ctx, email, code)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(data); err != nil {
			writeError(w, fmt.Errorf("%w: writing verify body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

func (c *Controller) handleResetPasswordRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		email := r.URL.Query().Get("email")
		login := r.URL.Query().Get("login")
		oldPassword := r.URL.Query().Get("oldPassword")
		newPassword := r.URL.Query().Get("oldPassword")
		err := c.auth.ResetPassword(ctx, login, email, oldPassword, newPassword)
		if err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

func parseUser(r *http.Request) (*model.User, error) {
	var user model.User
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("user body cannot be read", "error", err)
		return nil, fmt.Errorf("%w: %w", errInvalidArguments, err)
	}

	if err = json.Unmarshal(body, &user); err != nil {
		slog.Warn("user cannot be unmarshalled", "error", err)
		return nil, fmt.Errorf("%w: %w", errInvalidArguments, err)
	}

	if _, err := mail.ParseAddress(user.Email); err != nil {
		slog.Warn("user's email is incorrect", "error", err)
		return nil, fmt.Errorf("%w: %w", errInvalidArguments, err)
	}

	return &user, nil
}
