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
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func Test_ListPosterRepository_GetPosters(t *testing.T) {
	t.Parallel()

	listID := 1
	testListPosters := testListPostersFabric{}.Generate([]int{1, 2}, listID)

	tests := []struct {
		name        string
		queryErr    error
		listPosters []*model.ListPoster
		setScanMock bool
		scanErr     error
		wantErr     error
	}{
		{
			name:        "ok",
			queryErr:    nil,
			listPosters: testListPosters,
			setScanMock: true,
			scanErr:     nil,
			wantErr:     nil,
		},
		{
			name:        "query not found",
			setScanMock: false,
			scanErr:     nil,
			queryErr:    pgx.ErrNoRows,
			wantErr:     ErrNotFound,
		},
		{
			name:        "scan not found",
			setScanMock: true,
			queryErr:    nil,
			scanErr:     pgx.ErrNoRows,
			wantErr:     pgx.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			db := mocks.NewPostgresDBMock(t)
			db.QueryMock.Set(func(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
				if tt.queryErr != nil {
					return nil, tt.queryErr
				}

				rowsMock := mocks.NewRowsMock(t)
				rowsMock.CloseMock.Return()
				rowsMock.ErrMock.Return(nil)

				return rowsMock, nil
			})
			if tt.setScanMock {
				db.ScanAllMock.Set(func(dst interface{}, _ pgx.Rows) error {
					dstPtr, ok := dst.(*[]*model.ListPoster)
					if !ok {
						return fmt.Errorf("invalid dest type")
					}

					if tt.scanErr != nil {
						return tt.scanErr
					}

					*dstPtr = tt.listPosters
					return nil
				})
			}

			repo := NewListPosterRepository(db)

			// Act
			gotListPosters, gotErr := repo.GetPosters(ctx, listID)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}

			if diff := cmp.Diff(tt.listPosters, gotListPosters); diff != "" {
				t.Errorf("ListPosterRepository_GetPosters(%q): got unexpected diff (-want,+got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_ListPosterRepository_AddPosters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		count        int
		setExecMock  bool
		getErr       error
		execErr      error
		wantPosition int
		wantErr      error
	}{
		{
			name:         "ok",
			count:        5,
			setExecMock:  true,
			getErr:       nil,
			execErr:      nil,
			wantPosition: 6,
			wantErr:      nil,
		},
		{
			name:        "count err",
			count:       0,
			setExecMock: false,
			getErr:      assert.AnError,
			wantErr:     assert.AnError,
		},
		{
			name:        "exec err",
			count:       1,
			setExecMock: true,
			getErr:      nil,
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

				if tt.getErr != nil {
					return tt.getErr
				}

				destPtr, ok := dest.(*int)
				if !ok {
					return fmt.Errorf("invalid dest type")
				}

				*destPtr = tt.count
				return nil
			})
			if tt.setExecMock {
				db.ExecMock.Set(func(_ context.Context, _ string, args ...interface{}) (pgconn.CommandTag, error) {
					if len(args) == 0 {
						return nil, fmt.Errorf("no args provided in database method")
					}

					if tt.execErr != nil {
						return nil, tt.execErr
					}

					pos, ok := args[2].(int)
					if !ok {
						return nil, fmt.Errorf("database method arg is invalid")
					}

					assert.Equal(t, tt.wantPosition, pos)

					return nil, nil
				})
			}

			repo := NewListPosterRepository(db)

			// Act
			gotErr := repo.AddPoster(ctx, 1, 1)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}
		})
	}
}

func Test_ListPosterRepository_MovePoster(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		count      int
		txSuccess  bool
		argsToFail int
		wantErr    error
	}{
		{
			name:      "ok",
			count:     1,
			txSuccess: true,
			wantErr:   nil,
		},
		{
			name:       "tx fail",
			txSuccess:  false,
			argsToFail: 3,
			wantErr:    assert.AnError,
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

				_, ok := dest.(*int)
				if !ok {
					return fmt.Errorf("invalid dest type")
				}

				return nil
			})
			db.TxExecMock.Times(2).Set(func(_ context.Context, _ pgx.Tx, _ string, args ...interface{}) (pgconn.CommandTag, error) {
				if !tt.txSuccess && len(args) == tt.argsToFail {
					return nil, assert.AnError
				}

				return nil, nil
			})

			txMock := mocks.NewTXMock(t)
			if tt.txSuccess {
				txMock.CommitMock.Return(nil)
			} else {
				txMock.RollbackMock.Return(nil)
			}

			db.TxBeginMock.Return(txMock, nil)

			repo := NewListPosterRepository(db)

			// Act
			gotErr := repo.MovePoster(ctx, 1, 1, 1)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}
		})
	}
}

func Test_ListPosterRepository_ChangePosterPosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		count      int
		txSuccess  bool
		argsToFail int
		wantErr    error
	}{
		{
			name:      "ok",
			count:     1,
			txSuccess: true,
			wantErr:   nil,
		},
		{
			name:       "tx fail",
			txSuccess:  false,
			argsToFail: 3,
			wantErr:    assert.AnError,
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

				dstPtr, ok := dest.(*int)
				if !ok {
					return fmt.Errorf("invalid dest type")
				}

				*dstPtr = 1

				return nil
			})
			db.TxExecMock.Times(2).Set(func(_ context.Context, _ pgx.Tx, _ string, args ...interface{}) (pgconn.CommandTag, error) {
				if !tt.txSuccess && len(args) == tt.argsToFail {
					return nil, assert.AnError
				}

				return nil, nil
			})

			txMock := mocks.NewTXMock(t)
			if tt.txSuccess {
				txMock.CommitMock.Return(nil)
			} else {
				txMock.RollbackMock.Return(nil)
			}

			db.TxBeginMock.Return(txMock, nil)

			repo := NewListPosterRepository(db)

			// Act
			gotErr := repo.ChangePosterPosition(ctx, 1, 2, 3)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}
		})
	}
}

func Test_ListPosterRepository_DeletePoster(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		count       int
		setExecMock bool
		getErr      error
		execErr     error
		wantErr     error
	}{
		{
			name:        "ok",
			count:       1,
			setExecMock: true,
			execErr:     nil,
			wantErr:     nil,
		},
		{
			name:        "not found",
			count:       0,
			setExecMock: false,
			wantErr:     ErrNotFound,
		},
		{
			name:        "exec err",
			count:       1,
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

				*destPtr = tt.count
				return nil
			})
			if tt.setExecMock {
				db.ExecMock.Set(func(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
					return nil, tt.execErr
				})
			}

			repo := NewListPosterRepository(db)

			// Act
			gotErr := repo.DeletePoster(ctx, 1, 1)

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

type testListPostersFabric struct{}

func (testListPostersFabric) Generate(posterIDs []int, listID int) []*model.ListPoster {
	return lo.Map(posterIDs, func(posterID int, index int) *model.ListPoster {
		return &model.ListPoster{
			ID:       index + 1,
			PosterID: posterID,
			ListID:   listID,
			Position: index + 1,
		}
	})
}
