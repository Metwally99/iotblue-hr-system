package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
)

func translatePostgresError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			return apperror.ErrConflict
		case "23503":
			return apperror.ErrNotFound
		case "23P01":
			return apperror.ErrConflict
		}
	}
	return err
}
