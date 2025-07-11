//go:generate mockgen -source ./database.go -destination=./mocks/database.go -package=mock_database
package dbpostgres

import (
	"context"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

//nolint:typecheck
type DBops interface {
	GetPool(_ context.Context) *pgxpool.Pool
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	//nolint:typecheck
	ScanAll(dst interface{}, rows pgx.Rows) error
	//nolint:typecheck
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	//nolint:typecheck
	ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	//nolint:typecheck
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	TxBegin(ctx context.Context) (pgx.Tx, error)
	TxExec(ctx context.Context, tx pgx.Tx, sql string, args ...interface{}) (commandTag pgconn.CommandTag, err error)
	//nolint:typecheck
	TxQuery(ctx context.Context, tx pgx.Tx, sql string, args ...interface{}) (pgx.Rows, error)
}

type Database struct {
	cluster *pgxpool.Pool
}

//nolint:typecheck
func newDatabase(cluster *pgxpool.Pool) *Database {
	return &Database{cluster: cluster}
}

func (db Database) GetPool(_ context.Context) *pgxpool.Pool {
	return db.cluster
}

func (db Database) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, db.cluster, dest, query, args...)
}

func (db Database) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, db.cluster, dest, query, args...)
}

//nolint:typecheck
func (db Database) ScanAll(dst interface{}, rows pgx.Rows) error {
	return pgxscan.ScanAll(dst, rows)
}

func (db Database) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.cluster.Exec(ctx, query, args...)
}

//nolint:typecheck
func (db Database) ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.cluster.QueryRow(ctx, query, args...)
}

//nolint:typecheck
func (db Database) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.cluster.Query(ctx, sql, args...)
}

func (db Database) TxBegin(ctx context.Context) (pgx.Tx, error) {
	return db.cluster.Begin(ctx)
}

func (db Database) TxExec(ctx context.Context, tx pgx.Tx, sql string, args ...interface{}) (
	commandTag pgconn.CommandTag, err error,
) {
	return tx.Exec(ctx, sql, args...)
}

func (db Database) TxQuery(ctx context.Context, tx pgx.Tx, sql string, args ...interface{}) (pgx.Rows, error) {
	return tx.Query(ctx, sql, args...)
}
