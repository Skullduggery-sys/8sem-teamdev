package service

import (
	"context"
	"errors"
	"testing"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errFakeListPosterExists   = errors.New("such list-poster pair already exists")
	errFakeListPosterNotFound = errors.New("such list-poster pair does not exist")
)

type fakeListPosterRepo struct {
	list       map[int]model.List
	listPoster map[int]model.ListPoster

	listPosterIDIncrement int
}

func createFakeListPosterRepo() *fakeListPosterRepo {
	return &fakeListPosterRepo{
		list:       make(map[int]model.List),
		listPoster: make(map[int]model.ListPoster),

		listPosterIDIncrement: 0,
	}
}

func (r *fakeListPosterRepo) GetPosters(_ context.Context, listID int) ([]*model.ListPoster, error) {
	listPosters := make([]*model.ListPoster, 0)
	for _, listPoster := range r.listPoster {
		if listPoster.ListID == listID {
			listPoster := listPoster
			listPosters = append(listPosters, &listPoster)
		}
	}

	if len(listPosters) == 0 {
		return nil, errFakeListPosterNotFound
	}

	return listPosters, nil
}

func (r *fakeListPosterRepo) AddPoster(_ context.Context, listID, posterID int) error {
	listSize := 0
	for _, listPoster := range r.listPoster {
		if listPoster.ListID == listID {
			listSize++

			if listPoster.PosterID == posterID {
				return errFakeListPosterExists
			}
		}
	}

	r.listPoster[r.listPosterIDIncrement] = model.ListPoster{
		ID:       r.listPosterIDIncrement,
		ListID:   listID,
		PosterID: posterID,
		Position: listSize + 1,
	}
	r.listPosterIDIncrement++

	return nil
}

func (r *fakeListPosterRepo) DeletePoster(_ context.Context, listID, posterID int) error {
	ok := false
	for id, listPoster := range r.listPoster {
		if listPoster.ListID == listID && listPoster.PosterID == posterID {
			delete(r.listPoster, id)
			ok = true
			break
		}
	}

	if !ok {
		return errFakeListPosterNotFound
	}

	return nil
}

// TODO: implement & add test later. There is already integration test for this method.
func (r *fakeListPosterRepo) ChangePosterPosition(_ context.Context, _, _, _ int) error {
	return assert.AnError
}

// imitate transaction.
func (r *fakeListPosterRepo) MovePoster(ctx context.Context, curListID, newListID, posterID int) error {
	if err := r.DeletePoster(ctx, curListID, posterID); err != nil {
		return err
	}

	if err := r.AddPoster(ctx, newListID, posterID); err != nil {
		//nolint:errcheck // try to rollback 'transaction'
		r.AddPoster(ctx, curListID, newListID)
		return err
	}

	return nil
}

func TestFake_ListAddPoster_and_ListGetPosters(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := createFakeListPosterRepo()
	service := NewListPosterService(repo)
	want := []*model.ListPoster{
		{
			ID:       0,
			ListID:   testList1.ID,
			PosterID: testPoster.ID,
			Position: 1,
		},
	}

	// Act
	err := service.AddPoster(ctx, testList1.ID, testPoster.ID)

	// Assert
	require.ErrorIs(t, err, ErrCreated)

	gotPosters, gotErr := service.GetPosters(ctx, testList1.ID)
	require.NoError(t, gotErr)
	assert.Equal(t, want, gotPosters)
}

func TestFake_ListMovePoster(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := createFakeListPosterRepo()
	service := NewListPosterService(repo)
	want := []*model.ListPoster{
		{
			ID:       1,
			ListID:   testList2.ID,
			PosterID: testPoster.ID,
			Position: 1,
		},
	}

	err := service.AddPoster(ctx, testList1.ID, testPoster.ID)
	require.ErrorIs(t, err, ErrCreated)

	// Act
	err = service.MovePoster(ctx, testList1.ID, testList2.ID, testPoster.ID)

	// Assert
	require.NoError(t, err)

	_, errGetOld := service.GetPosters(ctx, testList1.ID)
	require.ErrorIs(t, errGetOld, errFakeListPosterNotFound)

	got, errGetNew := service.GetPosters(ctx, testList2.ID)
	require.NoError(t, errGetNew)
	assert.Equal(t, want, got)
}

func TestFake_ListDeletePoster(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := createFakeListPosterRepo()
	service := NewListPosterService(repo)

	err := service.AddPoster(ctx, testList1.ID, testPoster.ID)
	require.ErrorIs(t, err, ErrCreated)

	// Act
	err = service.DeletePoster(ctx, testList1.ID, testPoster.ID)

	// Assert
	require.NoError(t, err)
	_, errGet := service.GetPosters(ctx, testList1.ID)
	require.ErrorIs(t, errGet, errFakeListPosterNotFound)
}

func (*fakeListPosterRepo) GetListIDByPosterID(context.Context, int) (int, error) {
	return 0, repository.ErrNotFound
}
