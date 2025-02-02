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

var testPoster = model.Poster{
	Name:   "test",
	Year:   2024,
	Genres: []string{"g1", "g2", "g3"},
	Chrono: 90,
}

func TestPoster_Ok(t *testing.T) {
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		db.SetUp(t)
		adminID := db.CreateGenesisUser(ctx)
		defer db.TearDown()

		// Arrange
		testPoster.UserID = adminID
		poster := &testPoster
		repo := repository.NewPosterRepository(db.DB)

		// Act & Assert

		// create
		id, err := repo.Create(ctx, poster)
		require.NoError(t, err)
		poster.ID = id

		// make sure it's created
		gotPoster, gotErr := repo.Get(ctx, id)
		require.NoError(t, gotErr)
		if diff := cmpPoster(*poster, *gotPoster); diff != "" {
			t.Errorf("created poster is different from expected (-want, +got):\n%s", diff)
		}

		// update
		poster.Name = "upd"
		err = repo.Update(ctx, poster)
		require.NoError(t, err)

		// make sure it's updated
		gotPoster, gotErr = repo.Get(ctx, id)
		require.NoError(t, gotErr)
		if diff := cmpPoster(*poster, *gotPoster); diff != "" {
			t.Errorf("updated poster is different from expected (-want, +got):\n%s", diff)
		}

		// delete
		err = repo.Delete(ctx, poster.ID)
		require.NoError(t, err)

		// make sure it's deleted
		_, gotErr = repo.Get(ctx, id)
		require.ErrorIs(t, gotErr, repository.ErrNotFound)
	})
}

/* helpers */

func createPosterWithCheck(ctx context.Context, t *testing.T, repo *repository.PosterRepository, wantPoster model.Poster) int {
	posterID, err := repo.Create(ctx, &wantPoster)
	require.NoError(t, err)

	gotPoster, gotErr := repo.Get(ctx, posterID)
	require.NoError(t, gotErr)
	if diff := cmpPoster(wantPoster, *gotPoster); diff != "" {
		t.Errorf("created poster is different from expected (-want, +got):\n%s", diff)
	}

	return posterID
}

func cmpPoster(poster1 model.Poster, poster2 model.Poster) string {
	poster1.ID = poster2.ID
	poster1.CreatedAt = poster2.CreatedAt

	return cmp.Diff(poster1, poster2)
}
