package mysql

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
  "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func Test_sqlite_Migrate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS users").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = Migrate(db)
	assert.NoErrorf(t, err, "unexpected error during migration: %s", err)

	err = mock.ExpectationsWereMet()
	assert.NoErrorf(t, err, "there were unfulfilled expectations: %s", err)
}

func Test_sqlite_Migrate_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS users").
		WillReturnError(mysql.ErrInvalidConn)

	err = Migrate(db)
	require.Errorf(t, err, "unexpected error during migration: %s", err)

	err = mock.ExpectationsWereMet()
	require.NoErrorf(t, err, "there were unfulfilled expectations: %s", err)
}
