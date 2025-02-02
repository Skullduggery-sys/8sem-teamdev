package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres/mocks"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

func Test_PosterRepository_Get(t *testing.T) {
	t.Parallel()

	testPoster := testPostersFabric{}.Generate()
	mappedTestDAO := mapPosterDAO(testPoster)

	tests := []struct {
		name       string
		posterID   int
		wantPoster *model.Poster
		wantErr    error
	}{
		{
			name:       "ok",
			posterID:   1,
			wantPoster: mappedTestDAO,
			wantErr:    nil,
		},
		{
			name:       "not found",
			posterID:   404,
			wantPoster: nil,
			wantErr:    ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			db := mocks.NewPostgresDBMock(t)
			db.GetMock.Set(func(_ context.Context, dest interface{}, _ string, args ...interface{}) error {
				if len(args) == 0 {
					return fmt.Errorf("no args provided in database method")
				}

				destPtr, ok := dest.(*posterDAO)
				if !ok {
					return fmt.Errorf("invalid dest type")
				}

				posterID, ok := args[0].(int)
				if !ok {
					return fmt.Errorf("first database method arg is invalid")
				}

				if posterID != testPoster.ID {
					return pgx.ErrNoRows
				}

				*destPtr = *testPoster
				return nil
			})

			repo := NewPosterRepository(db)

			// Act
			gotPoster, gotErr := repo.Get(ctx, tt.posterID)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}

			if diff := cmp.Diff(tt.wantPoster, gotPoster); diff != "" {
				t.Errorf("PosterRepository_Get(%q): got unexpected diff (-want,+got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_PosterRepository_Create(t *testing.T) {
	t.Parallel()

	testPoster := testPostersFabric{}.Generate()
	mappedTestDAO := mapPosterDAO(testPoster)

	tests := []struct {
		name    string
		poster  *model.Poster
		rowMock *mocks.RowMock
		wantID  int
		wantErr error
	}{
		{
			name:   "ok",
			poster: mappedTestDAO,
			rowMock: mocks.NewRowMock(t).ScanMock.Set(func(dest ...interface{}) error {
				destPtr, ok := dest[0].(*int)
				if !ok {
					return fmt.Errorf("provided scan dest is invalid")
				}

				*destPtr = mappedTestDAO.ID
				return nil
			}),
			wantID:  mappedTestDAO.ID,
			wantErr: nil,
		},
		{
			name:   "scan error",
			poster: mappedTestDAO,
			rowMock: mocks.NewRowMock(t).ScanMock.Set(func(dest ...interface{}) error {
				return assert.AnError
			}),
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			db := mocks.NewPostgresDBMock(t)
			db.ExecQueryRowMock.Set(func(_ context.Context, _ string, _ ...interface{}) pgx.Row {
				return tt.rowMock
			})

			repo := NewPosterRepository(db)

			// Act
			gotID, gotErr := repo.Create(ctx, tt.poster)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}

			if diff := cmp.Diff(tt.wantID, gotID); diff != "" {
				t.Errorf("PosterRepository_Create(%q): got unexpected diff (-want,+got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_PosterRepository_Update(t *testing.T) {
	t.Parallel()

	testPoster := testPostersFabric{}.Generate()
	mappedTestDAO := mapPosterDAO(testPoster)

	tests := []struct {
		name    string
		poster  *model.Poster
		execErr error
		wantErr error
	}{
		{
			name:    "ok",
			poster:  mappedTestDAO,
			execErr: nil,
			wantErr: nil,
		},
		{
			name:    "not found",
			poster:  &model.Poster{Genres: []string{}},
			execErr: pgx.ErrNoRows,
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			db := mocks.NewPostgresDBMock(t)
			db.ExecMock.Set(func(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
				return nil, tt.execErr
			})

			repo := NewPosterRepository(db)

			// Act
			gotErr := repo.Update(ctx, tt.poster)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}
		})
	}
}

func Test_PosterRepository_Delete(t *testing.T) {
	t.Parallel()

	testPoster := testPostersFabric{}.Generate()

	tests := []struct {
		name        string
		posterID    int
		setExecMock bool
		execErr     error
		wantErr     error
	}{
		{
			name:        "ok",
			posterID:    1,
			setExecMock: true,
			execErr:     nil,
			wantErr:     nil,
		},
		{
			name:        "not found",
			posterID:    404,
			setExecMock: false,
			wantErr:     ErrNotFound,
		},
		{
			name:        "exec err",
			posterID:    1,
			setExecMock: true,
			execErr:     assert.AnError,
			wantErr:     assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			db := mocks.NewPostgresDBMock(t)
			db.GetMock.Set(func(_ context.Context, dest interface{}, _ string, args ...interface{}) error {
				if len(args) == 0 {
					return fmt.Errorf("no args provided in database method")
				}

				destPtr, ok := dest.(*int)
				if !ok {
					return fmt.Errorf("invalid dest type")
				}

				posterID, ok := args[0].(int)
				if !ok {
					return fmt.Errorf("first database method arg is invalid")
				}

				if posterID != testPoster.ID {
					return pgx.ErrNoRows
				}

				*destPtr = testPoster.ID
				return nil
			})
			if tt.setExecMock {
				db.ExecMock.Set(func(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
					return nil, tt.execErr
				})
			}

			repo := NewPosterRepository(db)

			// Act
			gotErr := repo.Delete(ctx, tt.posterID)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}
		})
	}
}

/* helpers */

type testPostersFabric struct{}

func (testPostersFabric) Generate() *posterDAO {
	return &posterDAO{
		ID:     1,
		Name:   "test-poster",
		Year:   2002,
		Genres: "genre1,genre2",
		Chrono: 90,
		UserID: 10,
	}
}
