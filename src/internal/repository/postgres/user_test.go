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

func Test_UserRepository_GetByLogin(t *testing.T) {
	t.Parallel()

	testUser := testUsersFabric{}.Generate()
	mappedTestDAO := mapUserDAO(testUser)

	tests := []struct {
		name     string
		login    string
		wantUser *model.User
		wantErr  error
	}{
		{
			name:     "ok",
			login:    "test-login",
			wantUser: mappedTestDAO,
			wantErr:  nil,
		},
		{
			name:     "not found",
			login:    "404",
			wantUser: nil,
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

				destPtr, ok := dest.(*userDAO)
				if !ok {
					return fmt.Errorf("invalid dest type")
				}

				login, ok := args[0].(string)
				if !ok {
					return fmt.Errorf("first database method arg is invalid")
				}

				if login != testUser.Login {
					return pgx.ErrNoRows
				}

				*destPtr = *testUser
				return nil
			})

			repo := NewUserRepository(db)

			// Act
			gotUser, gotErr := repo.GetByLogin(ctx, tt.login)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}

			if diff := cmp.Diff(tt.wantUser, gotUser); diff != "" {
				t.Errorf("UserRepository_GetByLogin(%q): got unexpected diff (-want,+got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_UserRepository_Create(t *testing.T) {
	t.Parallel()

	testUser := testUsersFabric{}.Generate()
	mappedTestDAO := mapUserDAO(testUser)

	tests := []struct {
		name    string
		user    *model.User
		rowMock *mocks.RowMock
		wantID  int
		wantErr error
	}{
		{
			name: "ok",
			user: mappedTestDAO,
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
			user: mappedTestDAO,
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

			repo := NewUserRepository(db)

			// Act
			gotID, gotErr := repo.Create(ctx, tt.user)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}

			if diff := cmp.Diff(tt.wantID, gotID); diff != "" {
				t.Errorf("UserRepository_Create(%q): got unexpected diff (-want,+got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_UserRepository_Delete(t *testing.T) {
	t.Parallel()

	testUser := testUsersFabric{}.Generate()

	tests := []struct {
		name        string
		userID      int
		setExecMock bool
		execErr     error
		wantErr     error
	}{
		{
			name:        "ok",
			userID:      1,
			setExecMock: true,
			execErr:     nil,
			wantErr:     nil,
		},
		{
			name:        "not found",
			userID:      404,
			setExecMock: false,
			wantErr:     ErrNotFound,
		},
		{
			name:        "exec err",
			userID:      1,
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

				userID, ok := args[0].(int)
				if !ok {
					return fmt.Errorf("first database method arg is invalid")
				}

				if userID != testUser.ID {
					return pgx.ErrNoRows
				}

				*destPtr = testUser.ID
				return nil
			})
			if tt.setExecMock {
				db.ExecMock.Set(func(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
					return nil, tt.execErr
				})
			}

			repo := NewUserRepository(db)

			// Act
			gotErr := repo.Delete(ctx, tt.userID)

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

type testUsersFabric struct{}

func (testUsersFabric) Generate() *userDAO {
	return &userDAO{
		ID:    1,
		Name:  "test-user",
		Login: "test-login",
	}
}
