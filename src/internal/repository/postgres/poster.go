package postgres

import (
	"context"
	"errors"

	dbpostgres "git.iu7.bmstu.ru/vai20u117/testing/src/internal/db/postgres"
	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	"github.com/jackc/pgx/v4"
)

type PosterRepository struct {
	db dbpostgres.DBops
}

func NewPosterRepository(db dbpostgres.DBops) *PosterRepository {
	return &PosterRepository{db: db}
}

func (r *PosterRepository) Get(ctx context.Context, posterID int) (*model.Poster, error) {
	queryName := "PosterRepository/Get"
	query := `select id,name,year,genres,chrono,user_id,created_at from poster where id = $1`

	dao := posterDAO{}

	err := r.db.Get(ctx, &dao, query, posterID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return nil, formatError(queryName, err)
	}

	return mapPosterDAO(&dao), nil
}

func (r *PosterRepository) Create(ctx context.Context, poster *model.Poster) (int, error) {
	queryName := "PosterRepository/Create"
	query := `insert into poster(name,genres,year,chrono,user_id) values($1,$2,$3,$4,$5) returning id`

	dao := reverseMapPosterDAO(poster)

	var id int
	err := r.db.ExecQueryRow(ctx, query,
		dao.Name, dao.Genres, dao.Year, dao.Chrono, dao.UserID).Scan(&id)
	if err != nil {
		return 0, formatError(queryName, err)
	}

	return id, nil
}

func (r *PosterRepository) Update(ctx context.Context, poster *model.Poster) error {
	queryName := "PosterRepository/Update"
	query := `
		update Poster
		set name = $2, year = $3, genres = $4, chrono = $5
		where id = $1`

	dao := reverseMapPosterDAO(poster)

	_, err := r.db.Exec(ctx, query, dao.ID, dao.Name, dao.Year, dao.Genres, dao.Chrono)
	if errors.Is(err, pgx.ErrNoRows) {
		return formatError(queryName, ErrNotFound)
	} else if err != nil {
		return formatError(queryName, err)
	}

	return nil
}

func (r *PosterRepository) Delete(ctx context.Context, posterID int) error {
	queryName := "PosterRepository/Delete"
	checkQueryName := "PosterRepository/Delete.exists"
	query := `delete from poster where id = $1`
	checkQuery := `select id from poster where id = $1`

	var id int
	err := r.db.Get(ctx, &id, checkQuery, posterID)
	if errors.Is(err, pgx.ErrNoRows) {
		return formatError(checkQueryName, ErrNotFound)
	} else if err != nil {
		return formatError(checkQueryName, err)
	}

	_, err = r.db.Exec(ctx, query, posterID)
	if err != nil {
		return formatError(queryName, err)
	}

	return nil
}
