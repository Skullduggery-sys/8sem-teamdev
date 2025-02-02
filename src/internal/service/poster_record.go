package service

import (
	"context"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
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
	return s.repo.GetUserRecords(ctx, userID)
}

func (s *PosterRecordService) CreateRecord(ctx context.Context, posterID, userID int) (int, error) {
	return s.repo.CreateRecord(ctx, posterID, userID)
}

func (s *PosterRecordService) DeleteRecord(ctx context.Context, posterID int) error {
	return s.repo.DeleteRecord(ctx, posterID)
}
