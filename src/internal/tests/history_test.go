//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
	"github.com/stretchr/testify/require"
)

func TestPosterRecord_Ok(t *testing.T) {
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		db.SetUp(t)
		adminID := db.CreateGenesisUser(ctx)
		defer db.TearDown()

		// Arrange
		poster := &testPoster
		poster.UserID = adminID
		repo := repository.NewPosterRecordRepository(db.DB)
		repoPoster := repository.NewPosterRepository(db.DB)

		// Act & Assert

		// create poster
		id, err := repoPoster.Create(ctx, poster)
		require.NoError(t, err)
		poster.ID = id

		// make sure it's created
		gotPoster, gotErr := repoPoster.Get(ctx, id)
		require.NoError(t, gotErr)
		if diff := cmpPoster(*poster, *gotPoster); diff != "" {
			t.Errorf("created poster is different from expected (-want, +got):\n%s", diff)
		}

		// create record
		_, err = repo.CreateRecord(ctx, poster.ID, adminID)
		require.NoError(t, err)

		// list all records
		records, gotErr := repo.GetUserRecords(ctx, adminID)
		require.NoError(t, gotErr)

		// check records
		require.Equalf(t, 1, len(records), "length of records is unexpected: %+v", records)
		require.Equalf(t, poster.ID, records[0].PosterID,
			"record is unexpected: got=%d,expected=%d", poster.ID, records[0].PosterID)

		// delete record
		err = repo.DeleteRecord(ctx, poster.ID)
		require.NoError(t, err)

		// make sure it's deleted
		_, gotErr = repo.GetUserRecords(ctx, adminID)
		require.ErrorIs(t, gotErr, repository.ErrNotFound)
	})
}
