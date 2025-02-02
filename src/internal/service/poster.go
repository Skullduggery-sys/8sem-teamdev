package service

import (
	"context"
	"errors"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
)

type posterRepository interface {
	Get(ctx context.Context, posterID int) (*model.Poster, error)
	Create(ctx context.Context, poster *model.Poster) (int, error)
	Update(ctx context.Context, poster *model.Poster) error
	Delete(ctx context.Context, posterID int) error
}

type PosterService struct {
	repo posterRepository
}

func NewPosterService(repo posterRepository) *PosterService {
	return &PosterService{repo: repo}
}

func (s *PosterService) Get(ctx context.Context, posterID int) (*model.Poster, error) {
	poster, err := s.repo.Get(ctx, posterID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return poster, nil
}

func (s *PosterService) Create(ctx context.Context, poster *model.Poster) (int, error) {
	return s.repo.Create(ctx, poster)
}

func (s *PosterService) Update(ctx context.Context, poster *model.Poster) error {
	err := s.repo.Update(ctx, poster)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func (s *PosterService) Delete(ctx context.Context, posterID int) error {
	err := s.repo.Delete(ctx, posterID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return nil
}
