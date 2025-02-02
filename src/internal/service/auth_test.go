package service

import (
	"context"
	"errors"
	"testing"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func Test_AuthService_GetUserID(t *testing.T) {
	t.Parallel()

	testToken := "test-token"

	tests := []struct {
		name       string
		token      string
		wantUserID int
		wantErr    error
	}{
		{
			name:       "ok",
			token:      testToken,
			wantUserID: 101,
			wantErr:    nil,
		},
		{
			name:    "not found",
			token:   "404",
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			service := NewAuthService(nil, "")
			service.sessions = map[string]*Session{
				testToken: {UserID: 101},
			}

			// Act
			gotUserID, gotErr := service.GetUserID(tt.token)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}

			assert.Equal(t, tt.wantUserID, gotUserID)
		})
	}
}

func Test_AuthService_GetUserTokenByAdmin(t *testing.T) {
	t.Parallel()

	testAdminSecret := "test-admin-secret"
	testLogin := "test-login"
	testUsers := []*model.User{
		{Login: testLogin},
	}

	tests := []struct {
		name        string
		adminSecret string
		login       string
		wantErr     error
	}{
		{
			name:        "ok",
			adminSecret: testAdminSecret,
			login:       testLogin,
			wantErr:     nil,
		},
		{
			name:        "unauthtorized",
			adminSecret: "bad-secret",
			wantErr:     ErrAdminIsNotAuthtorized,
		},
		{
			name:        "not found",
			adminSecret: testAdminSecret,
			login:       "404",
			wantErr:     ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			repo := &fakeUserRepo{users: testUsers}
			service := NewAuthService(repo, testAdminSecret)

			// Act
			gotToken, gotErr := service.GetUserTokenByAdmin(ctx, tt.adminSecret, tt.login)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)

				sessions := lo.Keys(service.sessions)
				assert.Equal(t, 1, len(sessions))
				assert.Equal(t, sessions[0], gotToken)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}
		})
	}
}

func Test_AuthService_SingUp(t *testing.T) {
	t.Parallel()

	testAdminSecret := "test-admin-secret"
	testLogin := "test-login"
	testUsers := []*model.User{
		{Login: testLogin},
	}

	tests := []struct {
		name    string
		user    *model.User
		wantErr error
	}{
		{
			name: "ok",
			user: &model.User{
				Login:    "new-login",
				Password: "test-password",
			},
			wantErr: nil,
		},
		{
			name: "bad admin",
			user: &model.User{
				Login:       "new-login",
				Role:        model.Admin.String(),
				AdminSecret: "bad-admin-secret",
				Password:    "test-password",
			},
			wantErr: ErrAdminIsNotAuthtorized,
		},
		{
			name: "user already exists",
			user: &model.User{
				Login:    testLogin,
				Password: "test-password",
			},
			wantErr: ErrLoginAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			repo := &fakeUserRepo{users: testUsers}
			service := NewAuthService(repo, testAdminSecret)

			// Act
			_, gotErr := service.SignUp(ctx, tt.user)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}
		})
	}
}

func Test_AuthService_SingIn(t *testing.T) {
	t.Parallel()

	testAdminSecret := "test-admin-secret"
	testLogin := "test-login"
	testPassword := "test-password"
	testUsers := []*model.User{
		{
			Login:    testLogin,
			Password: testPassword,
		},
	}

	tests := []struct {
		name    string
		user    *model.User
		wantErr error
	}{
		{
			name: "ok",
			user: &model.User{
				Login:    testLogin,
				Password: testPassword,
			},
			wantErr: nil,
		},
		{
			name: "user does not exist",
			user: &model.User{
				Login:    "new-login",
				Password: testPassword,
			},
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			repo := &fakeUserRepo{users: testUsers}
			service := NewAuthService(repo, testAdminSecret)

			// Act
			gotToken, gotErr := service.SignIn(ctx, tt.user)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)

				sessions := lo.Keys(service.sessions)
				assert.Equal(t, 1, len(sessions))
				assert.Equal(t, sessions[0], gotToken)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}
		})
	}
}

func Test_AuthService_SingOut(t *testing.T) {
	t.Parallel()

	testToken := "test-token"

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "ok",
			token:   testToken,
			wantErr: nil,
		},
		{
			name:    "not found",
			token:   "404",
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			service := NewAuthService(nil, "")
			service.sessions = map[string]*Session{
				testToken: {UserID: 101},
			}

			// Act
			gotErr := service.SignOut(ctx, tt.token)

			// Assert
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.True(t, errors.Is(gotErr, tt.wantErr))
			}
		})
	}
}

type fakeUserRepo struct {
	users []*model.User
}

func (*fakeUserRepo) Create(context.Context, *model.User) (int, error) {
	return 0, nil
}

func (f *fakeUserRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	filtered := lo.Filter(f.users, func(user *model.User, _ int) bool {
		if user.Login == login {
			return true
		}

		return false
	})
	if len(filtered) == 0 {
		return nil, repository.ErrNotFound
	}

	return filtered[0], nil
}
