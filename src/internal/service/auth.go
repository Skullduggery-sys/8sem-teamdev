package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
	"golang.org/x/crypto/bcrypt"
)

const passwordCost = 14

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (int, error)
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	GetByTGID(ctx context.Context, tgID string) (*model.User, error)
}

type AuthService struct {
	mx          sync.RWMutex
	sessions    map[string]*Session
	users       map[string]*model.User
	adminSecret string

	userRepo UserRepository
}

func NewAuthService(userRepo UserRepository, adminSecret string) *AuthService {
	return &AuthService{
		mx:          sync.RWMutex{},
		sessions:    make(map[string]*Session),
		users:       make(map[string]*model.User),
		adminSecret: adminSecret,
		userRepo:    userRepo,
	}
}

func (a *AuthService) GetUserByTGID(ctx context.Context, tgID string) (*model.User, error) {
	a.mx.RLock()
	if user, ok := a.users[tgID]; ok {
		a.mx.RUnlock()
		return user, nil
	}
	a.mx.RUnlock()

	user, err := a.userRepo.GetByTGID(ctx, tgID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	a.mx.Lock()
	a.users[tgID] = user
	a.mx.Unlock()

	return user, nil
}

func (a *AuthService) GetUserID(token string) (int, error) {
	session, ok := a.sessions[token]
	if !ok {
		return 0, ErrNotFound
	}

	return session.UserID, nil
}

func (a *AuthService) GetUserTokenByAdmin(ctx context.Context, adminSecret, login string) (string, error) {
	if adminSecret != a.adminSecret {
		return "", ErrAdminIsNotAuthtorized
	}

	userInDB, err := a.userRepo.GetByLogin(ctx, login)
	if errors.Is(err, repository.ErrNotFound) {
		return "", ErrNotFound
	} else if err != nil {
		return "", err
	}

	session, err := NewSession(userInDB.ID, model.DefaultUser.String())
	if err != nil {
		return "", fmt.Errorf("failed to create new sessions: %w", err)
	}

	a.mx.Lock()
	a.sessions[session.Token] = session
	a.mx.Unlock()

	return session.Token, nil
}

func (a *AuthService) SignUp(ctx context.Context, user *model.User) (int, error) {
	if user.Role == model.Admin.String() && user.AdminSecret != a.adminSecret {
		return 0, ErrAdminIsNotAuthtorized
	}

	if _, err := a.userRepo.GetByLogin(ctx, user.Login); err == nil {
		return 0, ErrLoginAlreadyExists
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), passwordCost)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrGeneratingHash, err)
	}

	user.Password = string(bytes)

	return a.userRepo.Create(ctx, user)
}

func (a *AuthService) SignIn(ctx context.Context, user *model.User) (string, error) {
	userInDB, err := a.userRepo.GetByLogin(ctx, user.Login)
	if errors.Is(err, repository.ErrNotFound) {
		return "", fmt.Errorf("%w: user with login %q not found", ErrNotFound, user.Login)
	} else if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userInDB.Password), []byte(user.Password)); err != nil {
		slog.Warn("password mismatch", "error", err)
		return "", fmt.Errorf("%w: password mismatch", ErrBadPassword)
	} else if userInDB.Role == model.Admin.String() && user.AdminSecret != a.adminSecret {
		return "", fmt.Errorf("%w bad admin auth", ErrAdminIsNotAuthtorized)
	}

	session, err := NewSession(userInDB.ID, userInDB.Role)
	if err != nil {
		return "", fmt.Errorf("failed to create new sessions: %w", err)
	}

	a.mx.Lock()
	a.sessions[session.Token] = session
	a.mx.Unlock()

	return session.Token, nil
}

func (a *AuthService) SignOut(ctx context.Context, token string) error {
	if _, ok := a.sessions[token]; !ok {
		return fmt.Errorf("%w: no session with such token, sign in first", ErrNotFound)
	}

	a.mx.Lock()
	delete(a.sessions, token)
	a.mx.Unlock()

	return nil
}

func (a *AuthService) Authorize(token string) error {
	if _, ok := a.sessions[token]; !ok {
		return fmt.Errorf("%w: no session with such token, sign in first", ErrNotFound)
	}

	return nil
}
