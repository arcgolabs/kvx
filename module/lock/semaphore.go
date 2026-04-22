package lock

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/arcgolabs/kvx"
	"github.com/samber/oops"
)

// Semaphore provides a distributed semaphore implementation.
type Semaphore struct {
	client kvx.KV
	key    string
	limit  int
}

// NewSemaphore creates a new Semaphore.
func NewSemaphore(client kvx.KV, key string, limit int) *Semaphore {
	return &Semaphore{
		client: client,
		key:    key,
		limit:  limit,
	}
}

// Acquire acquires a permit.
func (s *Semaphore) Acquire(ctx context.Context, ttl time.Duration) error {
	if s == nil {
		return oops.In("kvx/module/lock").
			With("op", "acquire_semaphore", "ttl", ttl).
			New("semaphore is nil")
	}
	if s.client == nil {
		return oops.In("kvx/module/lock").
			With("op", "acquire_semaphore", "key", s.key, "ttl", ttl, "limit", s.limit).
			New("lock client is nil")
	}
	count, err := s.loadCount(ctx, true)
	if err != nil {
		return err
	}
	if count >= s.limit {
		return oops.In("kvx/module/lock").
			With("op", "acquire_semaphore", "key", s.key, "ttl", ttl, "limit", s.limit, "count", count).
			Wrapf(ErrLockNotAcquired, "acquire semaphore")
	}

	return s.storeCount(ctx, count+1, ttl)
}

// Release releases a permit.
func (s *Semaphore) Release(ctx context.Context) error {
	if s == nil {
		return oops.In("kvx/module/lock").
			With("op", "release_semaphore").
			New("semaphore is nil")
	}
	if s.client == nil {
		return oops.In("kvx/module/lock").
			With("op", "release_semaphore", "key", s.key, "limit", s.limit).
			New("lock client is nil")
	}
	count, err := s.loadCount(ctx, false)
	if err != nil {
		return err
	}
	if count > 0 {
		count--
	}
	return s.storeCount(ctx, count, 0)
}

func (s *Semaphore) loadCount(ctx context.Context, allowMissing bool) (int, error) {
	data, err := s.client.Get(ctx, s.key)
	if err != nil {
		if allowMissing && errors.Is(err, kvx.ErrNil) {
			return 0, nil
		}
		return 0, oops.In("kvx/module/lock").
			With("op", "load_semaphore_count", "key", s.key, "allow_missing", allowMissing).
			Wrapf(err, "load semaphore count")
	}
	if len(data) == 0 {
		return 0, nil
	}

	count, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, oops.In("kvx/module/lock").
			With("op", "parse_semaphore_count", "key", s.key, "raw_count", string(data)).
			Wrapf(err, "parse semaphore count")
	}
	return count, nil
}

func (s *Semaphore) storeCount(ctx context.Context, count int, ttl time.Duration) error {
	value := []byte(strconv.Itoa(count))
	if err := s.client.Set(ctx, s.key, value, ttl); err != nil {
		return oops.In("kvx/module/lock").
			With("op", "store_semaphore_count", "key", s.key, "count", count, "ttl", ttl).
			Wrapf(err, "store semaphore count")
	}
	return nil
}
