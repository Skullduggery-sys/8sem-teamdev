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

func Test_PosterRecord_GetUserRecords(t *testing.T) {
	t.Parallel()

	testRecords := testRecrodsFabric{}.Generate([]int{1, 2}, 1)

	tests := []struct {
		name        string
		userID      int
		records     []*model.PosterRecord
		setScanMock bool
		scanErr     error
		queryErr    error
		wantErr     error
	}{
		{
			name:        "ok",
			userID:      1,
			records:     testRecords,
			setScanMock: true,
			scanErr:     nil,
			queryErr:    nil,
			wantErr:     nil,
		},
		{
			name:        "query not found",
			userID:      404,
			setScanMock: false,
			scanErr:     nil,
			queryErr:    pgx.ErrNoRows,
			wantErr:     ErrNotFound,
		},
		{
			name:        "scan not found",
			userID:      404,
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
					dstPtr, ok := dst.(*[]*model.PosterRecord)
					if !ok {
						return fmt.Errorf("invalid dest type")
					}

					if tt.scanErr != nil {
						return tt.scanErr
					}

					*dstPtr = tt.records
					return nil
				})
			}

			repo := NewPosterRecordRepository(db)

			// Act
			gotRecords, gotErr := repo.GetUserRecords(ctx, tt.userID)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}

			if diff := cmp.Diff(tt.records, gotRecords); diff != "" {
				t.Errorf("PosterRecord_GetUserRecords(%q): got unexpected diff (-want,+got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_PosterRecord_CreateRecord(t *testing.T) {
	t.Parallel()

	recordID := 1

	tests := []struct {
		name    string
		rowMock *mocks.RowMock
		wantID  int
		wantErr error
	}{
		{
			name: "ok",
			rowMock: mocks.NewRowMock(t).ScanMock.Set(func(dest ...interface{}) error {
				destPtr, ok := dest[0].(*int)
				if !ok {
					return fmt.Errorf("provided scan dest is invalid")
				}

				*destPtr = recordID
				return nil
			}),
			wantID:  recordID,
			wantErr: nil,
		},
		{
			name: "scan error",
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

			repo := NewPosterRecordRepository(db)

			// Act
			gotID, gotErr := repo.CreateRecord(ctx, 1, 1)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}

			if diff := cmp.Diff(tt.wantID, gotID); diff != "" {
				t.Errorf("PosterRecord_CreateRecord(%q): got unexpected diff (-want,+got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_PosterRecord_DeleteRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		execErr error
		wantErr error
	}{
		{
			name:    "ok",
			execErr: nil,
			wantErr: nil,
		},
		{
			name:    "exec err",
			execErr: assert.AnError,
			wantErr: assert.AnError,
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

			repo := NewPosterRecordRepository(db)

			// Act
			gotErr := repo.DeleteRecord(ctx, 1)

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

type testRecrodsFabric struct{}

func (testRecrodsFabric) Generate(posterIDs []int, userID int) []*model.PosterRecord {
	return lo.Map(posterIDs, func(posterID int, index int) *model.PosterRecord {
		return &model.PosterRecord{
			ID:       index + 1,
			PosterID: posterID,
			UserID:   userID,
		}
	})
}
