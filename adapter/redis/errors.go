package redis

import (
	"errors"

	"github.com/arcgolabs/kvx"
	goredis "github.com/redis/go-redis/v9"
	"github.com/samber/oops"
)

func wrapRedisError(op string, err error) error {
	if err == nil {
		return nil
	}

	return oops.In("kvx/adapter/redis").
		With("op", op).
		Wrapf(err, "redis %s", op)
}

func wrapRedisNilResult[T any](op string, value T, err error) (T, error) {
	if err == nil {
		return value, nil
	}

	var zero T
	if errors.Is(err, goredis.Nil) {
		return zero, oops.In("kvx/adapter/redis").
			With("op", op).
			Wrapf(kvx.ErrNil, "redis %s", op)
	}

	return zero, wrapRedisError(op, err)
}

func wrapRedisResult[T any](op string, value T, err error) (T, error) {
	if err == nil {
		return value, nil
	}

	var zero T
	return zero, wrapRedisError(op, err)
}
