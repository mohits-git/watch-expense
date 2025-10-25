package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/mohits-git/watch-expense/internal/adapters/mysql"
)

func SetupDB(ctx context.Context, dsn string) *sql.DB {
	db, err := mysql.Connect(ctx, dsn)
	if err != nil {
		log.Println("Failed to connect to database:", err)
		panic(err)
	}

	err = mysql.Migrate(db)
	if err != nil {
		log.Println("Failed to migrate the database:", err)
		panic(err)
	}
	log.Println("Database connected and migrated successfully")
	return db
}
