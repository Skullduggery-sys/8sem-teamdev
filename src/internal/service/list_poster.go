package service

import (
	"context"
	"errors"
	"log/slog"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
)

type listPosterRepository interface {
	GetPosters(ctx context.Context, listID int) ([]*model.ListPoster, error)
	AddPoster(ctx context.Context, listID, posterID int) error
	MovePoster(ctx context.Context, curListID, newListID, posterID int) error
	ChangePosterPosition(ctx context.Context, listID, posterID, newPosition int) error
	DeletePoster(ctx context.Context, listID, posterID int) error
	GetListIDByPosterID(ctx context.Context, posterID int) (int, error)
}

type ListPosterService struct {
	repo listPosterRepository
}

func NewListPosterService(repo listPosterRepository) *ListPosterService {
	return &ListPosterService{repo: repo}
}

func (s *ListPosterService) GetPosters(ctx context.Context, listID int) ([]*model.ListPoster, error) {
	listPosters, err := s.repo.GetPosters(ctx, listID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return listPosters, nil
}

func (s *ListPosterService) AddPoster(ctx context.Context, listID, posterID int) error {
	// 1. get cur id
	// 2. if pgx.ErrNoRows -> add
	// 3. perform movePoster

	curListID, findErr := s.repo.GetListIDByPosterID(ctx, posterID)
	switch {
	case errors.Is(findErr, repository.ErrNotFound):
		err := s.repo.AddPoster(ctx, listID, posterID)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		} else if err != nil {
			return err
		}

		return ErrCreated
	case findErr != nil:
		return findErr
	default:
		slog.Debug("moving poster...", "cur_list_id", curListID, "new_list_id", listID, "poster_id", posterID)
		return s.MovePoster(ctx, curListID, listID, posterID)
	}
}

func (s *ListPosterService) MovePoster(ctx context.Context, curListID, newListID, posterID int) error {
	err := s.repo.MovePoster(ctx, curListID, newListID, posterID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func (s *ListPosterService) ChangePosterPosition(ctx context.Context, listID, posterID, newPosition int) error {
	err := s.repo.ChangePosterPosition(ctx, listID, posterID, newPosition)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func (s *ListPosterService) DeletePoster(ctx context.Context, listID, posterID int) error {
	err := s.repo.DeletePoster(ctx, listID, posterID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return nil
}
