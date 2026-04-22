package lock

import (
	"context"
	"time"

	"github.com/arcgolabs/kvx"
	"github.com/samber/oops"
)

// RWLock provides a read-write lock implementation using Redis.
type RWLock struct {
	client     kvx.KV
	readKey    string
	writeKey   string
	identifier string
}

// NewRWLock creates a new RWLock.
func NewRWLock(client kvx.KV, key string) *RWLock {
	return &RWLock{
		client:     client,
		readKey:    key + ":read",
		writeKey:   key + ":write",
		identifier: generateIdentifier(),
	}
}

// RLock acquires a read lock.
func (rw *RWLock) RLock(ctx context.Context, ttl time.Duration) error {
	if rw == nil {
		return oops.In("kvx/module/lock").
			With("op", "acquire_read_lock", "ttl", ttl).
			New("rwlock is nil")
	}
	if rw.client == nil {
		return oops.In("kvx/module/lock").
			With("op", "acquire_read_lock", "read_key", rw.readKey, "write_key", rw.writeKey, "ttl", ttl).
			New("lock client is nil")
	}
	// Check if write lock is held.
	exists, err := rw.client.Exists(ctx, rw.writeKey)
	if err != nil {
		return oops.In("kvx/module/lock").
			With("op", "acquire_read_lock", "read_key", rw.readKey, "write_key", rw.writeKey, "ttl", ttl).
			Wrapf(err, "check write lock")
	}
	if exists {
		return oops.In("kvx/module/lock").
			With("op", "acquire_read_lock", "read_key", rw.readKey, "write_key", rw.writeKey, "ttl", ttl).
			Wrapf(ErrLockNotAcquired, "acquire read lock")
	}

	// This simplified implementation stores one key per reader.
	if err := rw.client.Set(ctx, rw.readerKey(), []byte("1"), ttl); err != nil {
		return oops.In("kvx/module/lock").
			With("op", "acquire_read_lock", "read_key", rw.readKey, "reader_key", rw.readerKey(), "write_key", rw.writeKey, "ttl", ttl).
			Wrapf(err, "store read lock")
	}
	return nil
}

// RUnlock releases a read lock.
func (rw *RWLock) RUnlock(ctx context.Context) error {
	if rw == nil {
		return oops.In("kvx/module/lock").
			With("op", "release_read_lock").
			New("rwlock is nil")
	}
	if rw.client == nil {
		return oops.In("kvx/module/lock").
			With("op", "release_read_lock", "read_key", rw.readKey, "reader_key", rw.readerKey()).
			New("lock client is nil")
	}
	if err := rw.client.Delete(ctx, rw.readerKey()); err != nil {
		return oops.In("kvx/module/lock").
			With("op", "release_read_lock", "read_key", rw.readKey, "reader_key", rw.readerKey()).
			Wrapf(err, "delete read lock")
	}
	return nil
}

// Lock acquires a write lock.
func (rw *RWLock) Lock(ctx context.Context, ttl time.Duration) error {
	if rw == nil {
		return oops.In("kvx/module/lock").
			With("op", "acquire_write_lock", "ttl", ttl).
			New("rwlock is nil")
	}
	if rw.client == nil {
		return oops.In("kvx/module/lock").
			With("op", "acquire_write_lock", "write_key", rw.writeKey, "ttl", ttl).
			New("lock client is nil")
	}
	if err := rw.client.Set(ctx, rw.writeKey, []byte(rw.identifier), ttl); err != nil {
		return oops.In("kvx/module/lock").
			With("op", "acquire_write_lock", "write_key", rw.writeKey, "ttl", ttl).
			Wrapf(err, "store write lock")
	}
	return nil
}

// Unlock releases a write lock.
func (rw *RWLock) Unlock(ctx context.Context) error {
	if rw == nil {
		return oops.In("kvx/module/lock").
			With("op", "release_write_lock").
			New("rwlock is nil")
	}
	if rw.client == nil {
		return oops.In("kvx/module/lock").
			With("op", "release_write_lock", "write_key", rw.writeKey).
			New("lock client is nil")
	}
	if err := rw.client.Delete(ctx, rw.writeKey); err != nil {
		return oops.In("kvx/module/lock").
			With("op", "release_write_lock", "write_key", rw.writeKey).
			Wrapf(err, "delete write lock")
	}
	return nil
}

func (rw *RWLock) readerKey() string {
	return rw.readKey + ":" + rw.identifier
}
