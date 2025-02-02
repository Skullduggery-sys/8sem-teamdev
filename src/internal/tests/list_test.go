//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
	cmp "github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

var (
	genesis = model.List{
		Name: "root",
	}
	testList1 = model.List{
		Name: "test list_1",
	}
	testList2 = model.List{
		Name: "test list_2",
	}
)

func TestList_Ok(t *testing.T) {
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		db.SetUp(t)
		adminID := db.CreateGenesisUser(ctx)
		genesis.ID = db.CreateGenesisList(ctx, adminID)
		testList1.ParentID = genesis.ID
		defer db.TearDown()

		// Arrange
		listCopy := testList1
		list := &listCopy
		list.UserID = adminID
		repo := repository.NewListRepository(db.DB)

		// Act & Assert

		// create
		id, err := repo.Create(ctx, list)
		require.NoError(t, err)
		list.ID = id

		// make sure it's created
		gotList, gotErr := repo.Get(ctx, id)
		require.NoError(t, gotErr)
		if diff := cmpList(*list, *gotList); diff != "" {
			t.Errorf("created list is different from expected (-want, +got):\n%s", diff)
		}

		// update
		list.Name = "upd"
		err = repo.Update(ctx, list)
		require.NoError(t, err)

		// make sure it's updated
		gotList, gotErr = repo.Get(ctx, id)
		require.NoError(t, gotErr)
		if diff := cmpList(*list, *gotList); diff != "" {
			t.Errorf("updated list is different from expected (-want, +got):\n%s", diff)
		}

		// delete
		err = repo.Delete(ctx, list.ID)
		require.NoError(t, err)

		// make sure it's deleted
		_, gotErr = repo.Get(ctx, id)
		require.ErrorIs(t, gotErr, repository.ErrNotFound)
	})
}

/* helpers */

func cmpList(list1 model.List, list2 model.List) string {
	list1.ID = list2.ID

	return cmp.Diff(list1, list2)
}
