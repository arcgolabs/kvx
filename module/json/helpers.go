package json

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/samber/oops"
)

func (j *JSON) getDocumentData(ctx context.Context, key string) ([]byte, error) {
	if j == nil || j.client == nil {
		return nil, oops.In("kvx/module/json").
			With("op", "get", "key", key, "path", "$").
			New("json client is nil")
	}
	data, err := j.client.JSONGet(ctx, key, "$")
	if err != nil {
		return nil, oops.In("kvx/module/json").
			With("op", "get", "key", key, "path", "$").
			Wrapf(err, "get json document")
	}
	return data, nil
}

func (j *JSON) getPathData(ctx context.Context, key, path string) ([]byte, error) {
	if j == nil || j.client == nil {
		return nil, oops.In("kvx/module/json").
			With("op", "get_path", "key", key, "path", path).
			New("json client is nil")
	}
	data, err := j.client.JSONGetField(ctx, key, path)
	if err != nil {
		return nil, oops.In("kvx/module/json").
			With("op", "get_path", "key", key, "path", path).
			Wrapf(err, "get json path")
	}
	return data, nil
}

func (j *JSON) setDocumentData(ctx context.Context, key string, data []byte, expiration time.Duration) error {
	if j == nil || j.client == nil {
		return oops.In("kvx/module/json").
			With("op", "set", "key", key, "path", "$", "expiration", expiration).
			New("json client is nil")
	}
	if err := j.client.JSONSet(ctx, key, "$", data, expiration); err != nil {
		return oops.In("kvx/module/json").
			With("op", "set", "key", key, "path", "$", "expiration", expiration).
			Wrapf(err, "set json document")
	}
	return nil
}

func (j *JSON) setPathData(ctx context.Context, key, path string, data []byte) error {
	if j == nil || j.client == nil {
		return oops.In("kvx/module/json").
			With("op", "set_path", "key", key, "path", path).
			New("json client is nil")
	}
	if err := j.client.JSONSetField(ctx, key, path, data); err != nil {
		return oops.In("kvx/module/json").
			With("op", "set_path", "key", key, "path", path).
			Wrapf(err, "set json path")
	}
	return nil
}

func (j *JSON) deletePath(ctx context.Context, key, path string) error {
	if j == nil || j.client == nil {
		return oops.In("kvx/module/json").
			With("op", "delete", "key", key, "path", path).
			New("json client is nil")
	}
	if err := j.client.JSONDelete(ctx, key, path); err != nil {
		return oops.In("kvx/module/json").
			With("op", "delete", "key", key, "path", path).
			Wrapf(err, "delete json path")
	}
	return nil
}

func marshalJSONValue(op string, value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, oops.In("kvx/module/json").
			With("op", op, "stage", "marshal", "value_type", reflect.TypeOf(value)).
			Wrapf(err, "marshal json value")
	}
	return data, nil
}

func unmarshalJSONValue(data []byte, dest any, op string) error {
	if err := json.Unmarshal(data, dest); err != nil {
		return oops.In("kvx/module/json").
			With("op", op, "stage", "unmarshal", "dest_type", reflect.TypeOf(dest)).
			Wrapf(err, "unmarshal json value")
	}
	return nil
}
