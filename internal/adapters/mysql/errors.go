package mysql

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

func HandleMysqlError(err error) error {
	if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows in result set") {
		return apperr.NewAppError(apperr.ErrNotFound, "record not found", err)
	}
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		switch mysqlErr.Number {
		case 1062:
			return apperr.NewAppError(apperr.ErrConflict, "unique constraint violation", err)
		case 1451, 1452:
			return apperr.NewAppError(apperr.ErrConflict, "foreign key constraint violation", err)
		case 1048:
			return apperr.NewAppError(apperr.ErrInvalid, "cannot be null constraint violation", err)
		}
	}
	return apperr.NewAppError(apperr.ErrInternal, "database error", err)
}
