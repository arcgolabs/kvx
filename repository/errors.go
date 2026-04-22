package repository

import "github.com/samber/oops"

// ErrNotFound reports that a repository entity does not exist.
var ErrNotFound = &repositoryError{"not found"}

// ErrOperationNotSupported reports that the selected backend cannot perform the requested operation.
var ErrOperationNotSupported = &repositoryError{"operation not supported"}

// ErrFieldNotFound reports that the requested entity field does not exist in repository metadata.
var ErrFieldNotFound = &repositoryError{"field not found"}

type repositoryError struct{ msg string }

func (e *repositoryError) Error() string { return "kvx: " + e.msg }

func wrapRepositoryError(err error, action string, fields ...any) error {
	if err != nil {
		args := append([]any{"action", action}, fields...)
		return oops.In("kvx/repository").
			With(args...).
			Wrapf(err, "%s", action)
	}

	return nil
}

func wrapRepositoryResult[T any](value T, err error, action string, fields ...any) (T, error) {
	if err != nil {
		var zero T
		args := append([]any{"action", action}, fields...)
		return zero, oops.In("kvx/repository").
			With(args...).
			Wrapf(err, "%s", action)
	}

	return value, nil
}
