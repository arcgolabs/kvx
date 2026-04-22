package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/arcgolabs/kvx"
)

// JSONSetField updates the top-level field selected by path.
func (b *JSONBackend) JSONSetField(_ context.Context, key, path string, value []byte) error {
	current, ok := b.data[key]
	if !ok {
		return kvx.ErrNil
	}

	var document map[string]any
	if err := json.Unmarshal(current, &document); err != nil {
		return fmt.Errorf("decode %s JSON document: %w", key, err)
	}

	var fieldValue any
	if err := json.Unmarshal(value, &fieldValue); err != nil {
		return fmt.Errorf("decode %s JSON field %q: %w", key, normalizeJSONFieldPath(path), err)
	}

	document[normalizeJSONFieldPath(path)] = fieldValue

	encoded, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("encode %s JSON document: %w", key, err)
	}

	b.data[key] = encoded
	return nil
}

// JSONGetField returns the top-level field selected by path.
func (b *JSONBackend) JSONGetField(_ context.Context, key, path string) ([]byte, error) {
	current, ok := b.data[key]
	if !ok {
		return nil, nil
	}

	var document map[string]json.RawMessage
	if err := json.Unmarshal(current, &document); err != nil {
		return nil, fmt.Errorf("decode %s JSON document: %w", key, err)
	}

	return document[normalizeJSONFieldPath(path)], nil
}

func normalizeJSONFieldPath(path string) string {
	return strings.TrimPrefix(path, "$.")
}
