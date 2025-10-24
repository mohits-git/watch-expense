package mysql

import (
	"database/sql"
	_ "embed"
	"log"
)

//go:embed schema.sql
var schemaSQL string

func Migrate(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	if err != nil {
		log.Println("Failed to execute migrations: ", err)
		return err
	}
	return nil
}
