package lock_test

import (
	"context"
	"errors"
	"testing"
	"time"

	lock "github.com/arcgolabs/kvx/module/lock"
)

type mockLockClient struct {
	owners map[string]string
}

func newMockLockClient() *mockLockClient {
	return &mockLockClient{owners: make(map[string]string)}
}

func (m *mockLockClient) Acquire(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	_ = ctx
	_ = ttl
	if _, exists := m.owners[key]; exists {
		return false, nil
	}
	m.owners[key] = token
	return true, nil
}

func (m *mockLockClient) Release(ctx context.Context, key, token string) (bool, error) {
	_ = ctx
	if owner, exists := m.owners[key]; exists && owner == token {
		delete(m.owners, key)
		return true, nil
	}
	return false, nil
}

func (m *mockLockClient) Extend(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	_ = ctx
	_ = ttl
	owner, exists := m.owners[key]
	return exists && owner == token, nil
}

func TestLock_UsesOwnershipTokenForRelease(t *testing.T) {
	ctx := context.Background()
	client := newMockLockClient()
	lock1 := lock.New(client, "resource", &lock.Options{TTL: time.Second, AutoExtend: false})
	lock2 := lock.New(client, "resource", &lock.Options{TTL: time.Second, AutoExtend: false})

	if err := lock1.Acquire(ctx); err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	if released, err := client.Release(ctx, "resource", "wrong-token"); err != nil {
		t.Fatalf("Release failed: %v", err)
	} else if released {
		t.Fatalf("wrong token should not release the lock")
	}

	if err := lock2.Acquire(ctx); !errors.Is(err, lock.ErrLockNotAcquired) {
		t.Fatalf("second lock should still be blocked, got %v", err)
	}

	if err := lock1.Release(ctx); err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	if err := lock2.Acquire(ctx); err != nil {
		t.Fatalf("second lock should acquire after owner release: %v", err)
	}
}

func TestLock_ExtendRequiresOwnershipToken(t *testing.T) {
	ctx := context.Background()
	client := newMockLockClient()
	distLock := lock.New(client, "resource", &lock.Options{TTL: time.Second, AutoExtend: false})

	if err := distLock.Acquire(ctx); err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	if ok, err := client.Extend(ctx, "resource", "wrong-token", time.Second); err != nil {
		t.Fatalf("Extend failed: %v", err)
	} else if ok {
		t.Fatalf("wrong token should not extend the lock")
	}

	if err := distLock.Extend(ctx, time.Second); err != nil {
		t.Fatalf("owner extend failed: %v", err)
	}
}
