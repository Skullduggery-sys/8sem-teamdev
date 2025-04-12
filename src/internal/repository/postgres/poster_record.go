package postgres

import (
	"context"
	"errors"
	"fmt"

	dbpostgres "git.iu7.bmstu.ru/vai20u117/testing/src/internal/db/postgres"
	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	"github.com/jackc/pgx/v4"
)

type PosterRecordRepository struct {
	db dbpostgres.DBops
}

func NewPosterRecordRepository(db dbpostgres.DBops) *PosterRecordRepository {
	return &PosterRecordRepository{db: db}
}

func (r *PosterRecordRepository) GetUserRecords(ctx context.Context, userID int) ([]*model.PosterRecord, error) {
	queryName := "PosterRecordRepository/GetUserRecords"
	query := `select id,poster_id,created_at from PosterRecord where user_id = $1`

	rows, err := r.db.Query(ctx, query, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return nil, formatError(queryName, err)
	} else if err = rows.Err(); err != nil {
		return nil, formatError(queryName, fmt.Errorf("rows: %w", err))
	}
	defer rows.Close()

	var records []*model.PosterRecord
	err = r.db.ScanAll(&records, rows)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return nil, formatError(queryName, err)
	} else if len(records) == 0 {
		return nil, formatError(queryName, ErrNotFound)
	}

	return records, nil
}

func (r *PosterRecordRepository) CreateRecord(ctx context.Context, posterID, userID int) (int, error) {
	queryName := "PosterRecordRepository/CreateRecord"
	query := `insert into PosterRecord (poster_id, user_id) values ($1,$2) returning id`

	var id int
	err := r.db.ExecQueryRow(ctx, query, posterID, userID).Scan(&id)
	if err != nil {
		return id, formatError(queryName, err)
	}

	return id, nil
}

func (r *PosterRecordRepository) DeleteRecord(ctx context.Context, posterID int) error {
	queryName := "PosterRecordRepository/DeleteRecord"
	query := `delete from PosterRecord where poster_id = $1`

	_, err := r.db.Exec(ctx, query, posterID)
	if err != nil {
		return formatError(queryName, err)
	}

	return nil
}
