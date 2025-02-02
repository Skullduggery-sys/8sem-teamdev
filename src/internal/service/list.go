package service

import (
	"context"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
)

type listRepository interface {
	Get(ctx context.Context, listID int) (*model.List, error)
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
	return s.repo.Get(ctx, listID)
}

func (s *ListService) Create(ctx context.Context, list *model.List) (int, error) {
	if list.ParentID == 0 {
		rootID, err := s.repo.GetRootID(ctx)
		if err != nil {
			return 0, err
		}

		list.ParentID = rootID
	}

	return s.repo.Create(ctx, list)
}

func (s *ListService) Update(ctx context.Context, list *model.List) error {
	return s.repo.Update(ctx, list)
}

func (s *ListService) Delete(ctx context.Context, listID int) error {
	// TODO change parentID of all posters that belong to this list to parentID of this list.
	return s.repo.Delete(ctx, listID)
}
