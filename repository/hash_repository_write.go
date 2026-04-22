package repository

import (
	"context"
	"errors"
	"time"

	"github.com/samber/lo"
)

type hashSaveState[T any] struct {
	key           string
	hashData      map[string][]byte
	removeEntries []string
	addEntries    []string
}

type hashFieldUpdateState struct {
	key           string
	storageField  string
	value         []byte
	removeEntries []string
	addEntries    []string
}

// Save stores an entity in the hash repository without setting a TTL.
func (r *HashRepository[T]) Save(ctx context.Context, entity *T) error {
	return r.SaveWithExpiration(ctx, entity, 0)
}

// SaveWithExpiration stores an entity in the hash repository and optionally sets a TTL.
func (r *HashRepository[T]) SaveWithExpiration(ctx context.Context, entity *T, expiration time.Duration) error {
	r.logDebug("kvx hash save started", "expiration_ms", expiration.Milliseconds())

	state, err := r.prepareHashSave(ctx, entity)
	if err != nil {
		r.logError("kvx hash save failed", "error", err)
		return err
	}
	if err := r.persistHashSave(ctx, state, expiration); err != nil {
		r.logError("kvx hash save failed", "key", state.key, "error", err)
		return err
	}

	r.logDebug("kvx hash save completed", "key", state.key, "indexed", len(state.addEntries))
	return nil
}

// SaveBatch stores a batch of entities in the hash repository without setting a TTL.
func (r *HashRepository[T]) SaveBatch(ctx context.Context, entities []*T) error {
	return r.SaveBatchWithExpiration(ctx, entities, 0)
}

// SaveBatchWithExpiration stores a batch of entities in the hash repository and optionally sets a TTL.
func (r *HashRepository[T]) SaveBatchWithExpiration(ctx context.Context, entities []*T, expiration time.Duration) error {
	if len(entities) == 0 {
		return nil
	}

	return runAll(entities, func(entity *T) error {
		return r.SaveWithExpiration(ctx, entity, expiration)
	})
}

// UpdateField updates a single stored hash field and refreshes any related index entries.
func (r *HashRepository[T]) UpdateField(ctx context.Context, id, fieldName string, value any) error {
	state, err := r.prepareHashFieldUpdate(ctx, id, fieldName, value)
	if err != nil {
		return err
	}

	if script, ok := r.script.Get(); ok {
		return execHashFieldUpdateScript(
			ctx,
			script,
			state.key,
			state.storageField,
			state.value,
			state.removeEntries,
			state.addEntries,
		)
	}

	if err := r.client.HSet(ctx, state.key, map[string][]byte{state.storageField: state.value}); err != nil {
		return wrapRepositoryError(err, "write hash field value", "op", "write_hash_field_value", "key", state.key, "field", state.storageField)
	}
	return r.base.indexer.ApplyIndexDiff(ctx, state.removeEntries, state.addEntries)
}

// IncrementField atomically increments a numeric hash field.
func (r *HashRepository[T]) IncrementField(ctx context.Context, id, fieldName string, increment int64) (int64, error) {
	metadata, err := r.base.metadataForType()
	if err != nil {
		return 0, err
	}

	_, fieldTag, exists := metadata.ResolveField(fieldName)
	if !exists {
		return 0, wrapRepositoryError(ErrFieldNotFound, "resolve hash field metadata", "op", "resolve_hash_field_metadata", "id", id, "field_name", fieldName)
	}

	result, incrErr := r.client.HIncrBy(ctx, r.base.keyFromID(id), fieldTag.StorageName(), increment)
	return wrapRepositoryResult(result, incrErr, "increment hash field", "op", "increment_hash_field", "id", id, "key", r.base.keyFromID(id), "field_name", fieldName, "storage_field", fieldTag.StorageName(), "increment", increment)
}

func (r *HashRepository[T]) prepareHashSave(ctx context.Context, entity *T) (hashSaveState[T], error) {
	metadata, err := r.base.metadata(entity)
	if err != nil {
		return hashSaveState[T]{}, err
	}

	key, err := r.base.keyBuilder.Build(entity, metadata)
	if err != nil {
		return hashSaveState[T]{}, wrapRepositoryError(err, "build hash entity key", "op", "build_hash_entity_key")
	}

	hashData, err := r.codec.Encode(entity, metadata)
	if err != nil {
		return hashSaveState[T]{}, wrapRepositoryError(err, "encode hash entity", "op", "encode_hash_entity", "key", key)
	}

	previous, _, err := r.loadPreviousHashEntity(ctx, key)
	if err != nil {
		return hashSaveState[T]{}, err
	}

	removeEntries, addEntries, err := r.base.indexer.ReplaceEntityIndexEntries(ctx, previous, entity, metadata, key)
	if err != nil {
		return hashSaveState[T]{}, wrapRepositoryError(err, "compute hash index diff", "op", "compute_hash_index_diff", "key", key)
	}

	return hashSaveState[T]{
		key:           key,
		hashData:      hashData,
		removeEntries: removeEntries,
		addEntries:    addEntries,
	}, nil
}

func (r *HashRepository[T]) loadPreviousHashEntity(ctx context.Context, key string) (*T, bool, error) {
	entity, err := r.findByKey(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, wrapRepositoryError(err, "load existing hash entity", "op", "load_existing_hash_entity", "key", key)
	}

	return entity, true, nil
}

func (r *HashRepository[T]) persistHashSave(ctx context.Context, state hashSaveState[T], expiration time.Duration) error {
	if script, ok := r.script.Get(); ok {
		return execHashUpsertScript(
			ctx,
			script,
			state.key,
			state.hashData,
			expiration,
			state.removeEntries,
			state.addEntries,
		)
	}

	if err := r.client.HSet(ctx, state.key, state.hashData); err != nil {
		return wrapRepositoryError(err, "write hash entity", "op", "write_hash_entity", "key", state.key, "field_count", len(state.hashData))
	}
	if expiration > 0 {
		if err := r.kv.Expire(ctx, state.key, expiration); err != nil {
			return wrapRepositoryError(err, "set hash entity expiration", "op", "expire_hash_entity", "key", state.key, "expiration", expiration)
		}
	}

	return r.base.indexer.ApplyIndexDiff(ctx, state.removeEntries, state.addEntries)
}

func (r *HashRepository[T]) prepareHashFieldUpdate(ctx context.Context, id, fieldName string, value any) (hashFieldUpdateState, error) {
	key := r.base.keyFromID(id)
	entity, err := r.findByKey(ctx, key)
	if err != nil {
		return hashFieldUpdateState{}, wrapRepositoryError(err, "load hash entity for field update", "op", "load_hash_entity_for_field_update", "id", id, "key", key, "field_name", fieldName)
	}

	metadata, err := r.base.metadata(entity)
	if err != nil {
		return hashFieldUpdateState{}, err
	}

	resolvedField, fieldTag, exists := metadata.ResolveField(fieldName)
	if !exists {
		return hashFieldUpdateState{}, wrapRepositoryError(ErrFieldNotFound, "resolve hash field metadata", "op", "resolve_hash_field_metadata", "id", id, "key", key, "field_name", fieldName)
	}

	encodedValue, err := r.codec.EncodeSingleValue(value)
	if err != nil {
		return hashFieldUpdateState{}, wrapRepositoryError(err, "encode hash field value", "op", "encode_hash_field_value", "id", id, "key", key, "field_name", fieldName)
	}

	removeEntries, addEntries, err := r.base.indexer.ReplaceFieldIndexEntries(metadata, resolvedField, key, entity, value)
	if err != nil {
		return hashFieldUpdateState{}, wrapRepositoryError(err, "compute hash field index diff", "op", "compute_hash_field_index_diff", "id", id, "key", key, "field_name", fieldName)
	}

	return hashFieldUpdateState{
		key:           key,
		storageField:  fieldTag.StorageName(),
		value:         encodedValue,
		removeEntries: removeEntries,
		addEntries:    addEntries,
	}, nil
}

func encodeHashData(data map[string][]byte) [][]byte {
	return lo.FlatMap(lo.Entries(data), func(entry lo.Entry[string, []byte], _ int) [][]byte {
		return [][]byte{[]byte(entry.Key), entry.Value}
	})
}
