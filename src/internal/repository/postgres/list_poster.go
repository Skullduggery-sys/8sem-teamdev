package postgres

import (
	"context"
	"errors"
	"fmt"

	dbpostgres "git.iu7.bmstu.ru/vai20u117/testing/src/internal/db/postgres"
	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	"github.com/jackc/pgx/v4"
)

type ListPosterRepository struct {
	db dbpostgres.DBops
}

func NewListPosterRepository(db dbpostgres.DBops) *ListPosterRepository {
	return &ListPosterRepository{db: db}
}

func (r *ListPosterRepository) GetPosters(ctx context.Context, listID int) ([]*model.ListPoster, error) {
	queryName := "ListPosterRepository/GetPosters"
	query := `select id,list_id,poster_id,position from listposter where list_id = $1 order by position`

	rows, err := r.db.Query(ctx, query, listID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return nil, formatError(queryName, err)
	} else if err = rows.Err(); err != nil {
		return nil, formatError(queryName, fmt.Errorf("rows: %w", err))
	}
	defer rows.Close()

	var dao []*model.ListPoster
	err = r.db.ScanAll(&dao, rows)
	if err != nil {
		return nil, formatError(queryName, err)
	} else if len(dao) == 0 {
		return nil, formatError(queryName, ErrNotFound)
	}

	return dao, nil
}

func (r *ListPosterRepository) AddPoster(ctx context.Context, listID, posterID int) error {
	queryName := "ListPosterRepository/AddPoster"
	//nolint:goconst // test maintanin convenience // TODO remove stupid nolints
	query := `insert into listposter(list_id,poster_id,position) values ($1,$2,$3)`
	//nolint:goconst // test maintanin convenience
	queryCount := `select count(*) from listposter where list_id = $1`

	// count posters in list
	var count int
	err := r.db.Get(ctx, &count, queryCount, listID)
	if err != nil {
		return formatError(queryName+".count", err)
	}

	// add poster in list
	position := count + 1
	_, err = r.db.Exec(ctx, query, listID, posterID, position)
	if err != nil {
		return formatError(queryName, err)
	}

	return nil
}

func (r *ListPosterRepository) MovePoster(ctx context.Context, curListID, newListID, posterID int) error {
	queryName := "ListPosterRepository/MovePoster"
	//nolint:goconst // test maintanin convenience
	queryDelete := `delete from listposter where list_id = $1 and poster_id = $2`
	queryCount := `select count(*) from listposter where list_id = $1`
	queryAdd := `insert into listposter(list_id,poster_id,position) values ($1,$2,$3)`

	// count posters in new list
	var count int
	err := r.db.Get(ctx, &count, queryCount, newListID)
	if err != nil {
		return formatError(queryName, err)
	}

	position := count + 1 // insert in the end of list

	tx, err := r.db.TxBegin(ctx)
	if err != nil {
		return formatError(queryName, fmt.Errorf("tx init: %w", err))
	}

	_, err = r.db.TxExec(ctx, tx, queryDelete, curListID, posterID)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		err = errors.Join(err, errRollback)

		return formatError(queryName, fmt.Errorf("deleting from old list: %w", err))
	}

	_, err = r.db.TxExec(ctx, tx, queryAdd, newListID, posterID, position)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		err = errors.Join(err, errRollback)

		return formatError(queryName, fmt.Errorf("adding poster in the new list: %w", err))
	}

	err = tx.Commit(ctx)
	if err != nil {
		return formatError(queryName, fmt.Errorf("committing tx: %w", err))
	}

	return nil
}

func (r *ListPosterRepository) ChangePosterPosition(ctx context.Context, listID, posterID, newPosition int) error {
	queryName := "ListPosterRepository/ChangePosterPosition"
	queryGetCurPos := `select position from listposter where list_id = $1 and poster_id = $2`
	querySetNewPos := `update listposter set position = $3 where list_id = $1 and poster_id = $2`
	queryBetweenPos := `
		update listposter
		set position = position + $4
		where
			list_id = $1 and
			position between $2 and $3`

	var curPosition int
	err := r.db.Get(ctx, &curPosition, queryGetCurPos, listID, posterID)
	if err != nil {
		return formatError(queryName+".countCurPosition", err)
	} else if curPosition == 0 {
		return formatError(queryName+".countCurPosition", ErrNotFound)
	}

	// if old pos < new: decrement (old, new], otherwise increment [new, old) positions
	startPos, endPos, increment := curPosition+1, newPosition, -1
	if curPosition > newPosition {
		startPos, endPos, increment = newPosition, curPosition-1, 1
	}

	tx, err := r.db.TxBegin(ctx)
	if err != nil {
		return formatError(queryName+".txbegin", fmt.Errorf("tx init: %w", err))
	}

	_, err = r.db.TxExec(ctx, tx, queryBetweenPos, listID, startPos, endPos, increment)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		err = errors.Join(err, errRollback)

		return formatError(queryName+".txChangePositions", fmt.Errorf("changing positions: %w", err))
	}

	_, err = r.db.TxExec(ctx, tx, querySetNewPos, listID, posterID, newPosition)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		err = errors.Join(err, errRollback)

		return formatError(queryName+".txChangePositions", fmt.Errorf("setting new pos: %w", err))
	}

	err = tx.Commit(ctx)
	if err != nil {
		return formatError(queryName, fmt.Errorf("committing tx: %w", err))
	}

	return nil
}

func (r *ListPosterRepository) DeletePoster(ctx context.Context, listID, posterID int) error {
	queryName := "ListPosterRepository/DeletePoster"
	checkQuery := `select count(*) from listposter where list_id = $1 and poster_id = $2`
	query := `delete from listposter where list_id = $1 and poster_id = $2`

	var count int
	err := r.db.Get(ctx, &count, checkQuery, listID, posterID)
	if err != nil {
		return formatError(queryName, err)
	} else if count == 0 {
		return formatError(queryName, ErrNotFound)
	}

	_, err = r.db.Exec(ctx, query, listID, posterID)
	if err != nil {
		return formatError(queryName, err)
	}

	return nil
}
