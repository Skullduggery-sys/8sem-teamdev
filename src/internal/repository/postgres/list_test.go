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

func Test_ListRepository_Get(t *testing.T) {
	t.Parallel()

	testList := testListsFabric{}.Generate()
	mappedTestDAO := mapListDAO(testList)

	tests := []struct {
		name     string
		listID   int
		wantList *model.List
		wantErr  error
	}{
		{
			name:     "ok",
			listID:   11,
			wantList: mappedTestDAO,
			wantErr:  nil,
		},
		{
			name:     "not found",
			listID:   404,
			wantList: nil,
			wantErr:  ErrNotFound,
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

				destPtr, ok := dest.(*listDAO)
				if !ok {
					return fmt.Errorf("invalid dest type")
				}

				listID, ok := args[0].(int)
				if !ok {
					return fmt.Errorf("first database method arg is invalid")
				}

				if listID != testList.ID {
					return pgx.ErrNoRows
				}

				*destPtr = *testList
				return nil
			})

			repo := NewListRepository(db)

			// Act
			gotList, gotErr := repo.Get(ctx, tt.listID)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}

			if diff := cmp.Diff(tt.wantList, gotList); diff != "" {
				t.Errorf("ListRepository_Get(%q): got unexpected diff (-want,+got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_ListRepository_GetRootID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rootID  int
		getErr  error
		wantErr error
	}{
		{
			name:    "ok",
			rootID:  1,
			wantErr: nil,
		},
		{
			name:    "not found",
			getErr:  pgx.ErrNoRows,
			wantErr: pgx.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			db := mocks.NewPostgresDBMock(t)
			db.GetMock.Set(func(_ context.Context, dest interface{}, _ string, _ ...interface{}) error {
				destPtr, ok := dest.(*int)
				if !ok {
					return fmt.Errorf("invalid dest type")
				}

				if tt.getErr != nil {
					return tt.getErr
				}

				*destPtr = tt.rootID
				return nil
			})

			repo := NewListRepository(db)

			// Act
			gotRootID, gotErr := repo.GetRootID(ctx)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}

			assert.Equal(t, tt.rootID, gotRootID)
		})
	}
}

func Test_ListRepository_Create(t *testing.T) {
	t.Parallel()

	testList := testListsFabric{}.Generate()
	mappedTestDAO := mapListDAO(testList)

	tests := []struct {
		name    string
		list    *model.List
		rowMock *mocks.RowMock
		wantID  int
		wantErr error
	}{
		{
			name: "ok",
			list: mappedTestDAO,
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
			name: "scan error",
			list: mappedTestDAO,
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

			repo := NewListRepository(db)

			// Act
			gotID, gotErr := repo.Create(ctx, tt.list)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}

			if diff := cmp.Diff(tt.wantID, gotID); diff != "" {
				t.Errorf("ListRepository_Create(%q): got unexpected diff (-want,+got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_ListRepository_Update(t *testing.T) {
	t.Parallel()

	testList := testListsFabric{}.Generate()
	mappedTestDAO := mapListDAO(testList)

	tests := []struct {
		name    string
		list    *model.List
		execErr error
		wantErr error
	}{
		{
			name:    "ok",
			list:    mappedTestDAO,
			execErr: nil,
			wantErr: nil,
		},
		{
			name:    "not found",
			list:    &model.List{},
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

			repo := NewListRepository(db)

			// Act
			gotErr := repo.Update(ctx, tt.list)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}
		})
	}
}

func Test_ListRepository_Delete(t *testing.T) {
	t.Parallel()

	testList := testListsFabric{}.Generate()

	tests := []struct {
		name        string
		listID      int
		count       int
		setExecMock bool
		execErr     error
		wantErr     error
	}{
		{
			name:        "ok",
			listID:      testList.ID,
			count:       1,
			setExecMock: true,
			execErr:     nil,
			wantErr:     nil,
		},
		{
			name:        "not found",
			listID:      404,
			count:       0,
			setExecMock: false,
			wantErr:     ErrNotFound,
		},
		{
			name:        "exec err",
			listID:      testList.ID,
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

				listID, ok := args[0].(int)
				if !ok {
					return fmt.Errorf("first database method arg is invalid")
				}

				if listID != testList.ID {
					return pgx.ErrNoRows
				}

				*destPtr = tt.count
				return nil
			})
			if tt.setExecMock {
				db.ExecMock.Set(func(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
					return nil, tt.execErr
				})
			}

			repo := NewListRepository(db)

			// Act
			gotErr := repo.Delete(ctx, tt.listID)

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

type testListsFabric struct{}

func (testListsFabric) Generate() *listDAO {
	return &listDAO{
		ID:       11,
		ParentID: 1,
		Name:     "test-list",
		UserID:   10,
	}
}
