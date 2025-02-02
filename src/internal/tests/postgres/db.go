package postgres

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	dbpostgres "git.iu7.bmstu.ru/vai20u117/testing/src/internal/db/postgres"
	"github.com/spf13/viper"
)

//nolint:gocritic // (embed mutex) testDB
type TDB struct {
	DB dbpostgres.DBops
	sync.Mutex
}

func NewFromEnv() *TDB {
	db, err := dbpostgres.NewDB(context.Background(), &dbpostgres.DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		panic(err)
	}

	return &TDB{DB: db}
}

func (d *TDB) SetUp(t *testing.T, args ...interface{}) {
	t.Helper()
	d.Lock()
	d.Truncate(context.Background())
}

func (d *TDB) TearDown() {
	defer d.Unlock()
	d.Truncate(context.Background())
}

func (d *TDB) Truncate(ctx context.Context) {
	var tables []string
	err := d.DB.Select(ctx, &tables, `
		select table_name from information_schema.tables
		where table_schema = 'public'
			and table_type = 'BASE TABLE'
			and table_name != 'goose_db_version'`)
	if err != nil {
		panic(err)
	}
	if len(tables) == 0 {
		panic("run migration plz")
	}

	q := "Truncate table " + strings.Join(tables, ",")
	if _, err := d.DB.Exec(ctx, q); err != nil {
		panic(err)
	}
}

func (d *TDB) CreateGenesisList(ctx context.Context, userID int) int {
	query := `insert into list(id,parent_id,name,user_id,is_root) values(default,null,'root',$1,true) returning id`

	var id int
	if err := d.DB.ExecQueryRow(ctx, query, userID).Scan(&id); err != nil {
		panic("creating genesis list: " + err.Error())
	}
	return id
}

func (d *TDB) CreateGenesisUser(ctx context.Context) int {
	query := `insert into appuser(id,name,login,role,password) values(1,'test','test','admin','test') returning id`

	var id int
	if err := d.DB.ExecQueryRow(ctx, query).Scan(&id); err != nil {
		panic("creating genesis list: " + err.Error())
	}
	return id
}
