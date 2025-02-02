//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
)

func TestListPoster_Ok(t *testing.T) {
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		db.SetUp(t)
		adminID := db.CreateGenesisUser(ctx)
		genesis.ID = db.CreateGenesisList(ctx, adminID)
		testList1.ParentID = genesis.ID
		defer db.TearDown()

		// Arrange
		list1 := &testList1
		list2 := &testList2
		poster := &testPoster
		poster2 := *poster
		poster3 := *poster
		repoList := repository.NewListRepository(db.DB)
		repoListPoster := repository.NewListPosterRepository(db.DB)
		repoPoster := repository.NewPosterRepository(db.DB)

		// Act & Assert

		// create hierarchical lists
		id1, err := repoList.Create(ctx, list1)
		require.NoError(t, err)
		list1.ID = id1
		list2.ParentID = id1

		id2, err := repoList.Create(ctx, list2)
		require.NoError(t, err)
		list2.ID = id2

		// make sure they are created
		gotList1, gotErr := repoList.Get(ctx, id1)
		require.NoError(t, gotErr)
		if diff := cmpList(*list1, *gotList1); diff != "" {
			t.Errorf("created list1 is different from expected (-want, +got):\n%s", diff)
		}
		gotList2, gotErr := repoList.Get(ctx, id2)
		require.NoError(t, gotErr)
		if diff := cmpList(*list2, *gotList2); diff != "" {
			t.Errorf("created list2 is different from expected (-want, +got):\n%s", diff)
		}

		// create posters
		posterID1 := createPosterWithCheck(ctx, t, repoPoster, *poster)
		posterID2 := createPosterWithCheck(ctx, t, repoPoster, poster2)
		posterID3 := createPosterWithCheck(ctx, t, repoPoster, poster3)

		// create poster in list1
		err = repoListPoster.AddPoster(ctx, list1.ID, posterID1)
		require.NoError(t, gotErr)

		// make sure it's created in list1
		gotPosters, gotErr := repoListPoster.GetPosters(ctx, list1.ID)
		require.NoError(t, gotErr)
		wantListPosters := []*model.ListPoster{{ID: gotPosters[0].ID, ListID: list1.ID, PosterID: posterID1, Position: 1}}
		require.ElementsMatch(t, wantListPosters, gotPosters)

		// move poster to list2
		err = repoListPoster.MovePoster(ctx, list1.ID, list2.ID, posterID1)
		require.NoError(t, err)

		// make sure it has removed from list1
		gotPosters, gotErr = repoListPoster.GetPosters(ctx, list1.ID)
		require.ErrorIs(t, gotErr, repository.ErrNotFound)

		// make sure it has moved to list2
		gotPosters, gotErr = repoListPoster.GetPosters(ctx, list2.ID)
		require.NoError(t, gotErr)
		wantListPosters = []*model.ListPoster{{ID: gotPosters[0].ID, ListID: list2.ID, PosterID: posterID1, Position: 1}}
		require.ElementsMatch(t, wantListPosters, gotPosters)

		// delete poster from list2
		err = repoListPoster.DeletePoster(ctx, list2.ID, posterID1)
		require.NoError(t, err)

		gotPosters, gotErr = repoListPoster.GetPosters(ctx, list2.ID)
		require.ErrorIs(t, gotErr, repository.ErrNotFound)

		// Add posters to list1
		err = repoListPoster.AddPoster(ctx, list1.ID, posterID1)
		require.NoError(t, err)
		err = repoListPoster.AddPoster(ctx, list1.ID, posterID2)
		require.NoError(t, err)
		err = repoListPoster.AddPoster(ctx, list1.ID, posterID3)
		require.NoError(t, err)

		// make sure that positions are in the same order of being inserted
		gotPosters, gotErr = repoListPoster.GetPosters(ctx, list1.ID)
		require.NoError(t, err)
		wantListPosters = []*model.ListPoster{
			{ListID: list1.ID, PosterID: posterID1, Position: 1},
			{ListID: list1.ID, PosterID: posterID2, Position: 2},
			{ListID: list1.ID, PosterID: posterID3, Position: 3},
		}
		if diff := cmp.Diff(wantListPosters, gotPosters, cmpopts.IgnoreFields(model.ListPoster{}, "ID")); diff != "" {
			t.Errorf("listPoster is different from expected (-want, +got):\n%s", diff)
		}

		// make last poster be the first and make sure that positions are set right
		err = repoListPoster.ChangePosterPosition(ctx, list1.ID, posterID3, 1)
		require.NoError(t, err)

		gotPosters, gotErr = repoListPoster.GetPosters(ctx, list1.ID)
		require.NoError(t, err)
		wantListPosters = []*model.ListPoster{
			{ListID: list1.ID, PosterID: posterID3, Position: 1},
			{ListID: list1.ID, PosterID: posterID1, Position: 2},
			{ListID: list1.ID, PosterID: posterID2, Position: 3},
		}
		if diff := cmp.Diff(wantListPosters, gotPosters, cmpopts.IgnoreFields(model.ListPoster{}, "ID")); diff != "" {
			t.Errorf("listPoster is different from expected (-want, +got):\n%s", diff)
		}

		// make second poster be the last and make sure that positions are set right
		err = repoListPoster.ChangePosterPosition(ctx, list1.ID, posterID1, 3)
		require.NoError(t, err)

		gotPosters, gotErr = repoListPoster.GetPosters(ctx, list1.ID)
		require.NoError(t, err)
		wantListPosters = []*model.ListPoster{
			{ListID: list1.ID, PosterID: posterID3, Position: 1},
			{ListID: list1.ID, PosterID: posterID2, Position: 2},
			{ListID: list1.ID, PosterID: posterID1, Position: 3},
		}
		if diff := cmp.Diff(wantListPosters, gotPosters, cmpopts.IgnoreFields(model.ListPoster{}, "ID")); diff != "" {
			t.Errorf("listPoster is different from expected (-want, +got):\n%s", diff)
		}
	})
}
