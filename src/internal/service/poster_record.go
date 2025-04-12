package service

import (
	"context"
	"errors"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
)

type PosterRecordRepository interface {
	GetUserRecords(ctx context.Context, userID int) ([]*model.PosterRecord, error)
	CreateRecord(ctx context.Context, posterID, userID int) (int, error)
	DeleteRecord(ctx context.Context, posterID int) error
}

type PosterRecordService struct {
	repo PosterRecordRepository
}

func NewPosterRecordService(repo PosterRecordRepository) *PosterRecordService {
	return &PosterRecordService{repo: repo}
}

func (s *PosterRecordService) GetUserRecords(ctx context.Context, userID int) ([]*model.PosterRecord, error) {
	records, err := s.repo.GetUserRecords(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return records, nil
}

func (s *PosterRecordService) CreateRecord(ctx context.Context, posterID, userID int) (int, error) {
	id, err := s.repo.CreateRecord(ctx, posterID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *PosterRecordService) DeleteRecord(ctx context.Context, posterID int) error {
	err := s.repo.DeleteRecord(ctx, posterID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return nil
}
