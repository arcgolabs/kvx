package stream

import (
	"context"
	"encoding/json"

	"github.com/samber/oops"
)

func wrapError(err error, action string) error {
	if err != nil {
		return oops.In("kvx/module/stream").
			With("action", action).
			Wrapf(err, "%s", action)
	}

	return nil
}

func wrapResult[T any](value T, err error, action string) (T, error) {
	if err != nil {
		var zero T
		return zero, oops.In("kvx/module/stream").
			With("action", action).
			Wrapf(err, "%s", action)
	}

	return value, nil
}

func wrapContextError(ctx context.Context, action string) error {
	return oops.In("kvx/module/stream").
		With("action", action).
		Wrapf(ctx.Err(), "%s", action)
}

func marshalJSON(v any, action string) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, oops.In("kvx/module/stream").
			With("action", action).
			Wrapf(err, "%s", action)
	}

	return data, nil
}
