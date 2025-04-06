package service

import (
	"context"
	"errors"
	"testing"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errFakeListNotFound = errors.New("list not found")

//nolint:gochecknoglobals // test variable
var (
	testList1 = model.List{
		ID:       10,
		ParentID: 1,
		Name:     "test list_1",
	}
	testList2 = model.List{
		ID:       15,
		ParentID: 1,
		Name:     "test list_2",
	}
)

type fakeListRepo struct {
	list map[int]model.List
}

func (r *fakeListRepo) Create(_ context.Context, list *model.List) (int, error) {
	r.list[list.ID] = *list
	return list.ID, nil
}

func (r *fakeListRepo) Get(_ context.Context, listID int) (*model.List, error) {
	if list, ok := r.list[listID]; ok {
		return &list, nil
	}

	return &model.List{}, errFakeListNotFound
}

func (r *fakeListRepo) Update(_ context.Context, list *model.List) error {
	if _, ok := r.list[list.ID]; ok {
		r.list[list.ID] = *list
		return nil
	}

	return errFakeListNotFound
}

func (r *fakeListRepo) Delete(_ context.Context, listID int) error {
	if _, ok := r.list[listID]; ok {
		delete(r.list, listID)
		return nil
	}

	return errFakeListNotFound
}

func (r *fakeListRepo) GetRootID(_ context.Context) (int, error) {
	return 0, nil
}

func (*fakeListRepo) GetSublists(context.Context, int) ([]*model.List, error) {
	return nil, nil
}

func (*fakeListRepo) GetUserRoot(context.Context, int, int) (*model.List, error) {
	return nil, nil
}

func createFakeListRepo() *fakeListRepo {
	return &fakeListRepo{
		list: make(map[int]model.List),
	}
}

func TestFake_ListCreate(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := createFakeListRepo()
	service := NewListService(repo)

	list := testList1

	// Act
	id, err := service.Create(ctx, &list)

	// Assert
	require.NoError(t, err)

	gotList, gotErr := service.repo.Get(ctx, id)
	require.NoError(t, gotErr)
	assert.Equal(t, list, *gotList)
}

func TestFake_ListUpdate_ok(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := createFakeListRepo()
	service := NewListService(repo)

	list := testList1

	id, err := repo.Create(ctx, &list)
	require.NoError(t, err)

	list.Name = "changed name"
	list.ParentID = 101010

	// Act
	err = service.Update(ctx, &list)

	// Assert
	require.NoError(t, err)

	gotList, gotErr := service.repo.Get(ctx, id)
	require.NoError(t, gotErr)
	assert.Equal(t, list, *gotList)
}

func TestFake_ListUpdate_notFound(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := createFakeListRepo()
	service := NewListService(repo)

	list := testList1

	// Act
	err := service.Update(ctx, &list)

	// Assert
	require.ErrorIs(t, err, errFakeListNotFound)
}

func TestFake_ListDelete_ok(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := createFakeListRepo()
	service := NewListService(repo)

	list := testList1

	id, err := repo.Create(ctx, &list)
	require.NoError(t, err)

	// Act
	err = service.Delete(ctx, id)

	// Assert
	require.NoError(t, err)

	_, gotErr := service.repo.Get(ctx, id)
	require.ErrorIs(t, gotErr, errFakeListNotFound)
}

func TestFake_ListDelete_notFound(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := createFakeListRepo()
	service := NewListService(repo)

	// Act
	err := service.Delete(ctx, testList1.ID)

	// Assert
	require.ErrorIs(t, err, errFakeListNotFound)
}
