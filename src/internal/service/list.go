package service

import (
	"context"
	"errors"
	"fmt"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
)

type listRepository interface {
	Get(ctx context.Context, listID int) (*model.List, error)
	GetSublists(ctx context.Context, listID int) ([]*model.List, error)
	GetUserRoot(ctx context.Context, globalRootID, userID int) (*model.List, error)
	GetRootID(ctx context.Context) (int, error)
	Create(ctx context.Context, list *model.List) (int, error)
	Update(ctx context.Context, list *model.List) error
	Delete(ctx context.Context, listID int) error
}

type ListService struct {
	repo listRepository
}

func NewListService(repo listRepository) *ListService {
	return &ListService{repo: repo}
}

func (s *ListService) Get(ctx context.Context, listID int) (*model.List, error) {
	list, err := s.repo.Get(ctx, listID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return list, nil
}

func (s *ListService) GetUserRoot(ctx context.Context, userID int) (*model.List, error) {
	globalRootID, err := s.repo.GetRootID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get id of global list root: %w", err)
	}

	root, err := s.repo.GetUserRoot(ctx, globalRootID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return root, nil
}

func (s *ListService) GetSublists(ctx context.Context, listID int) ([]*model.List, error) {
	return s.repo.GetSublists(ctx, listID)
}

func (s *ListService) Create(ctx context.Context, list *model.List) (int, error) {
	if list.ParentID == 0 {
		rootID, err := s.repo.GetRootID(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to get id of global list root: %w", err)
		}

		list.ParentID = rootID
	}

	return s.repo.Create(ctx, list)
}

func (s *ListService) Update(ctx context.Context, list *model.List) error {
	err := s.repo.Update(ctx, list)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func (s *ListService) Delete(ctx context.Context, listID int) error {
	err := s.repo.Delete(ctx, listID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return nil
}
