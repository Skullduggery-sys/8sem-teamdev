package service

import (
	"context"
	"fmt"
	"testing"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals // test variable
var testPoster = model.Poster{
	ID:     101,
	Name:   "someMovie",
	Year:   2024,
	Genres: []string{"genre1", "genre2", "genre3"},
	Chrono: 100,
}

type fakePosterRepo struct {
	repo map[int]model.Poster
}

func (r *fakePosterRepo) Create(_ context.Context, poster *model.Poster) (int, error) {
	poster.ID = len(r.repo) + 101
	r.repo[poster.ID] = *poster
	return poster.ID, nil
}

func (*fakePosterRepo) CreateFromKP(_ context.Context, _ string, _ int) (int, error) {
	return 0, fmt.Errorf("unimplemented")
}

func (r *fakePosterRepo) Get(_ context.Context, posterID int) (*model.Poster, error) {
	if poster, ok := r.repo[posterID]; ok {
		return &poster, nil
	}

	return &model.Poster{}, ErrNotFound
}

func (r *fakePosterRepo) Update(_ context.Context, poster *model.Poster) error {
	if _, ok := r.repo[poster.ID]; ok {
		r.repo[poster.ID] = *poster
		return nil
	}

	return ErrNotFound
}

func (r *fakePosterRepo) Delete(_ context.Context, posterID int) error {
	if _, ok := r.repo[posterID]; ok {
		delete(r.repo, posterID)
		return nil
	}

	return ErrNotFound
}

func TestFake_PosterCreate(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := &fakePosterRepo{
		repo: make(map[int]model.Poster),
	}
	service := NewPosterService(repo, nil, "")

	poster := testPoster

	// Act
	id, err := service.Create(ctx, &poster)

	// Assert
	require.NoError(t, err)

	gotPoster, gotErr := service.repo.Get(ctx, id)
	require.NoError(t, gotErr)
	assert.Equal(t, poster, *gotPoster)
}

func TestFake_PosterUpdate_ok(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := &fakePosterRepo{
		repo: make(map[int]model.Poster),
	}
	service := NewPosterService(repo, nil, "")

	poster := testPoster

	id, err := service.Create(ctx, &poster)
	require.NoError(t, err)

	poster.Chrono = 75

	// Act
	err = service.Update(ctx, &poster)

	// Assert
	require.NoError(t, err)

	gotPoster, gotErr := service.repo.Get(ctx, id)
	require.NoError(t, gotErr)
	assert.Equal(t, poster, *gotPoster)
}

func TestFake_PosterUpdate_notFound(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := &fakePosterRepo{
		repo: make(map[int]model.Poster),
	}
	service := NewPosterService(repo, nil, "")

	poster := testPoster

	// Act
	err := service.Update(ctx, &poster)

	// Assert
	require.ErrorIs(t, err, ErrNotFound)
}

func TestFake_PosterDelete_ok(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := &fakePosterRepo{
		repo: make(map[int]model.Poster),
	}
	service := NewPosterService(repo, nil, "")

	poster := testPoster

	id, err := service.Create(ctx, &poster)
	require.NoError(t, err)

	// Act
	err = service.Delete(ctx, id)

	// Assert
	require.NoError(t, err)

	_, gotErr := service.repo.Get(ctx, id)
	require.ErrorIs(t, gotErr, ErrNotFound)
}

func TestFake_PosterDelete_notFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := &fakePosterRepo{
		repo: make(map[int]model.Poster),
	}
	service := NewPosterService(repo, nil, "")

	// Act
	err := service.Delete(ctx, testPoster.ID)

	// Assert
	require.ErrorIs(t, err, ErrNotFound)
}
