package service

import (
	"context"
	"errors"
	"sync"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (int, error)
	GetByTGID(ctx context.Context, tgID string) (*model.User, error)
}

type AuthService struct {
	mx    sync.RWMutex
	users map[string]*model.User

	userRepo UserRepository
}

func NewAuthService(userRepo UserRepository, adminSecret string) *AuthService {
	return &AuthService{
		mx:       sync.RWMutex{},
		users:    make(map[string]*model.User),
		userRepo: userRepo,
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

func (a *AuthService) SignUp(ctx context.Context, tgID string) (int, error) {
	if _, err := a.userRepo.GetByTGID(ctx, tgID); err == nil {
		return 0, ErrLoginAlreadyExists
	}

	user := &model.User{TGID: tgID}
	return a.userRepo.Create(ctx, user)
}
