package postgres

import (
	"context"
	"errors"

	dbpostgres "git.iu7.bmstu.ru/vai20u117/testing/src/internal/db/postgres"
	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	"github.com/jackc/pgx/v4"
)

type UserRepository struct {
	db dbpostgres.DBops
}

func NewUserRepository(db dbpostgres.DBops) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	queryName := "UserRepository/GetByLogin"
	query := `select id,name,login,email,role,password from appuser where login = $1`

	dao := userDAO{}

	err := r.db.Get(ctx, &dao, query, login)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return nil, formatError(queryName, err)
	}

	return mapUserDAO(&dao), nil
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) (int, error) {
	queryName := "UserRepository/Create"
	query := `insert into appuser(name,login,email,role,password) values($1,$2,$3,$4,$5) returning id`

	dao := reverseMapUserDAO(user)

	var id int
	err := r.db.ExecQueryRow(ctx, query,
		dao.Name, dao.Login, dao.Email, dao.Role, dao.Password).Scan(&id)
	if err != nil {
		return id, formatError(queryName, err)
	}

	return id, nil
}

func (r *UserRepository) Delete(ctx context.Context, userID int) error {
	queryName := "UserRepository/Delete"
	checkQueryName := "UserRepository/Delete.exists"
	query := `delete from appuser where id = $1`
	checkQuery := `select id from appuser where id = $1`

	var id int
	err := r.db.Get(ctx, &id, checkQuery, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return formatError(checkQueryName, ErrNotFound)
	} else if err != nil {
		return formatError(checkQueryName, err)
	}

	_, err = r.db.Exec(ctx, query, userID)
	if err != nil {
		return formatError(queryName, err)
	}

	return nil
}

func (r *UserRepository) ResetPassword(ctx context.Context, login, newPassword string) error {
	queryName := "UserRepository/ResetPassword"
	query := `
		update appuser
		set password = $2
		where login = $1`

	_, err := r.db.Exec(ctx, query, login, newPassword)
	if errors.Is(err, pgx.ErrNoRows) {
		return formatError(queryName, ErrNotFound)
	} else if err != nil {
		return formatError(queryName, err)
	}

	return nil
}
