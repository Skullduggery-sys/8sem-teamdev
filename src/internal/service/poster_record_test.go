package service

import (
	"context"
	"testing"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO map[int]bool -> map[int]model.
type fakePosterRecordRepo struct {
	repo map[int]bool
}

func (r *fakePosterRecordRepo) GetUserRecords(_ context.Context, _ int) ([]*model.PosterRecord, error) {
	return nil, nil
}

func (r *fakePosterRecordRepo) CreateRecord(_ context.Context, posterID, _ int) (int, error) {
	r.repo[posterID] = true
	return 0, nil
}

func (r *fakePosterRecordRepo) DeleteRecord(_ context.Context, posterID int) error {
	if _, ok := r.repo[posterID]; !ok {
		return ErrNotFound
	}

	delete(r.repo, posterID)
	return nil
}

func TestFake_PosterRecordCreate(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := &fakePosterRecordRepo{
		repo: make(map[int]bool),
	}
	service := NewPosterRecordService(repo)

	// Act
	_, err := service.CreateRecord(ctx, testPoster.ID, testPoster.UserID)

	// Assert
	require.NoError(t, err)
	assert.True(t, repo.repo[testPoster.ID])
}

func TestFake_PosterRecordDelete(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := &fakePosterRecordRepo{
		repo: make(map[int]bool),
	}
	service := NewPosterRecordService(repo)

	_, err := service.CreateRecord(ctx, testPoster.ID, testPoster.UserID)
	require.NoError(t, err)

	// Act
	err = service.DeleteRecord(ctx, testPoster.ID)

	// Assert
	require.NoError(t, err)
	assert.False(t, repo.repo[testPoster.ID])
}
