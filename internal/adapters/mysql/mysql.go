package mysql

import (
	"context"
	"database/sql"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var once sync.Once

func Connect(ctx context.Context, dsn string) (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	once.Do(func() {
		var err error
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		db.SetMaxOpenConns(1)
		_, err = db.ExecContext(ctx, "PRAGMA foreign_keys = ON;")
		if err != nil {
			panic(err)
		}
	})

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
