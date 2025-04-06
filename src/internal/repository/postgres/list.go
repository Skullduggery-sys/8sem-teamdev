package postgres

import (
	"context"
	"errors"
	"fmt"

	dbpostgres "git.iu7.bmstu.ru/vai20u117/testing/src/internal/db/postgres"
	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	"github.com/jackc/pgx/v4"
)

type ListRepository struct {
	db dbpostgres.DBops
}

func NewListRepository(db dbpostgres.DBops) *ListRepository {
	return &ListRepository{db: db}
}

func (r *ListRepository) Get(ctx context.Context, listID int) (*model.List, error) {
	queryName := "ListRepository/Get"
	query := `select id,name,user_id,parent_id from list where id = $1`

	dao := listDAO{}

	err := r.db.Get(ctx, &dao, query, listID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return nil, formatError(queryName, err)
	}

	return mapListDAO(&dao), err
}

func (r *ListRepository) GetUserRoot(ctx context.Context, globalRootID, userID int) (*model.List, error) {
	queryName := "ListRepository/GetUserRoot"
	query := `select id,name,user_id,parent_id from list where parent_id = $1 and user_id = $2 limit 1`

	dao := listDAO{}

	err := r.db.Get(ctx, &dao, query, globalRootID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return nil, formatError(queryName, err)
	}

	return mapListDAO(&dao), err
}

func (r *ListRepository) GetSublists(ctx context.Context, listID int) ([]*model.List, error) {
	queryName := "ListRepository/GetSublists"
	query := `select id,name,user_id,parent_id from list where parent_id = $1`

	rows, err := r.db.Query(ctx, query, listID)
	if errors.Is(err, pgx.ErrNoRows) {
		return []*model.List{}, nil
	} else if err != nil {
		return nil, formatError(queryName, err)
	} else if err = rows.Err(); err != nil {
		return nil, formatError(queryName, fmt.Errorf("rows: %w", err))
	}
	defer rows.Close()

	// TODO: use dao instead
	var dao []*model.List
	err = r.db.ScanAll(&dao, rows)
	if errors.Is(err, pgx.ErrNoRows) || len(dao) == 0 {
		return []*model.List{}, nil
	} else if err != nil {
		return nil, formatError(queryName, err)
	}

	return dao, nil
}

func (r *ListRepository) GetRootID(ctx context.Context) (int, error) {
	queryName := "ListRepository/GetRootID"
	query := `select id from list where is_root = true`

	var rootID int
	err := r.db.Get(ctx, &rootID, query)
	if err != nil {
		return 0, formatError(queryName, err)
	}

	return rootID, nil
}

func (r *ListRepository) Create(ctx context.Context, list *model.List) (int, error) {
	queryName := "ListRepository/Create"
	query := `insert into list(parent_id,name,user_id) values($1,$2,$3) returning id`

	dao := reverseMapListDAO(list)

	var id int
	err := r.db.ExecQueryRow(ctx, query, dao.ParentID, dao.Name, dao.UserID).Scan(&id)
	if err != nil {
		return id, formatError(queryName, err)
	}

	return id, nil
}

func (r *ListRepository) Update(ctx context.Context, list *model.List) error {
	queryName := "ListRepository/Update"
	query := `
		update list
		set name = $2
		where id = $1`

	dao := reverseMapListDAO(list)

	_, err := r.db.Exec(ctx, query, dao.ID, dao.Name)
	if err != nil {
		return formatError(queryName, err)
	}

	return nil
}

func (r *ListRepository) Delete(ctx context.Context, listID int) error {
	queryName := "ListRepository/Delete"
	checkQuery := `select count(*) from list where id = $1`
	query := `delete from list where id = $1`

	var count int
	err := r.db.Get(ctx, &count, checkQuery, listID)
	if err != nil {
		return formatError(queryName, err)
	} else if count == 0 {
		return formatError(queryName, ErrNotFound)
	}

	_, err = r.db.Exec(ctx, query, listID)
	if err != nil {
		return formatError(queryName, err)
	}

	return nil
}
