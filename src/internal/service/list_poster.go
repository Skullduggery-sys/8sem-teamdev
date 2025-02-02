package service

import (
	"context"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
)

type listPosterRepository interface {
	GetPosters(ctx context.Context, listID int) ([]*model.ListPoster, error)
	AddPoster(ctx context.Context, listID, posterID int) error
	MovePoster(ctx context.Context, curListID, newListID, posterID int) error
	ChangePosterPosition(ctx context.Context, listID, posterID, newPosition int) error
	DeletePoster(ctx context.Context, listID, posterID int) error
}

type ListPosterService struct {
	repo listPosterRepository
}

func NewListPosterService(repo listPosterRepository) *ListPosterService {
	return &ListPosterService{repo: repo}
}

func (s *ListPosterService) GetPosters(ctx context.Context, listID int) ([]*model.ListPoster, error) {
	return s.repo.GetPosters(ctx, listID)
}

func (s *ListPosterService) AddPoster(ctx context.Context, listID, posterID int) error {
	return s.repo.AddPoster(ctx, listID, posterID)
}

func (s *ListPosterService) MovePoster(ctx context.Context, curListID, newListID, posterID int) error {
	return s.repo.MovePoster(ctx, curListID, newListID, posterID)
}

func (s *ListPosterService) ChangePosterPosition(ctx context.Context, listID, posterID, newPosition int) error {
	return s.repo.ChangePosterPosition(ctx, listID, posterID, newPosition)
}

func (s *ListPosterService) DeletePoster(ctx context.Context, listID, posterID int) error {
	return s.repo.DeletePoster(ctx, listID, posterID)
}
