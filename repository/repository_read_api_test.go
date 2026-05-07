package repository_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/arcgolabs/kvx/repository"
)

func TestHashRepository_TryFindByID(t *testing.T) {
	ctx := context.Background()
	hash := newMockHash()
	kv := newMockKV()
	repo := repository.NewHashRepository[TestUser](hash, kv, "user")

	hash.data["user:user1"] = map[string][]byte{
		"name":  []byte("John Doe"),
		"email": []byte("john@example.com"),
		"age":   []byte("30"),
	}

	user, found, err := repo.TryFindByID(ctx, "user1")
	if err != nil {
		t.Fatalf("TryFindByID failed: %v", err)
	}
	if !found || user == nil || user.Email != "john@example.com" {
		t.Fatalf("expected to find user1, found=%v user=%#v", found, user)
	}

	user, found, err = repo.TryFindByID(ctx, "missing")
	if err != nil {
		t.Fatalf("TryFindByID missing failed: %v", err)
	}
	if found || user != nil {
		t.Fatalf("expected missing user to return nil,false; got %#v,%v", user, found)
	}
}

func TestHashRepository_FindFirstByField(t *testing.T) {
	ctx := context.Background()
	hash := newMockHash()
	kv := newMockKV()
	repo := repository.NewHashRepository[TestUser](hash, kv, "user")

	user := &TestUser{ID: "user1", Name: "John Doe", Email: "john@example.com", Age: 30}
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := repo.FindFirstByField(ctx, "Email", "john@example.com")
	if err != nil {
		t.Fatalf("FindFirstByField failed: %v", err)
	}
	if got.ID != "user1" {
		t.Fatalf("expected user1, got %s", got.ID)
	}

	_, err = repo.FindFirstByField(ctx, "Email", "missing@example.com")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestHashRepository_CountAndFindFirstIgnoreIndexKeys(t *testing.T) {
	ctx := context.Background()
	hash := newMockHash()
	kv := newMockKV()
	repo := repository.NewHashRepository[TestUser](hash, kv, "user")

	kv.data["user:user1"] = []byte("exists")
	kv.data["user:idx:email:john@example.com:user1"] = []byte("1")
	hash.data["user:user1"] = map[string][]byte{
		"name":  []byte("John Doe"),
		"email": []byte("john@example.com"),
		"age":   []byte("30"),
	}

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected index keys to be ignored, got count %d", count)
	}

	got, err := repo.FindFirst(ctx)
	if err != nil {
		t.Fatalf("FindFirst failed: %v", err)
	}
	if got.ID != "user1" {
		t.Fatalf("expected user1, got %s", got.ID)
	}
}

func TestHashRepository_FindFirstByFields(t *testing.T) {
	ctx := context.Background()
	hash := newMockHash()
	kv := newMockKV()
	repo := repository.NewHashRepository[TestUser](hash, kv, "user")

	users := []*TestUser{
		{ID: "user1", Name: "John Doe", Email: "john@example.com", Age: 30},
		{ID: "user2", Name: "Jane Doe", Email: "jane@example.com", Age: 30},
	}
	for _, user := range users {
		if err := repo.Save(ctx, user); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	got, err := repo.FindFirstByFields(ctx, map[string]string{
		"Email": "jane@example.com",
		"Age":   "30",
	})
	if err != nil {
		t.Fatalf("FindFirstByFields failed: %v", err)
	}
	if got.ID != "user2" {
		t.Fatalf("expected user2, got %s", got.ID)
	}
}

func TestJSONRepository_TryFindByID(t *testing.T) {
	ctx := context.Background()
	client := newMockJSON()
	kv := newMockKV()
	repo := repository.NewJSONRepository[TestUser](client, kv, "user")

	payload, err := json.Marshal(&TestUser{ID: "user1", Name: "John Doe", Email: "john@example.com", Age: 30})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	client.data["user:user1"] = payload

	user, found, err := repo.TryFindByID(ctx, "user1")
	if err != nil {
		t.Fatalf("TryFindByID failed: %v", err)
	}
	if !found || user == nil || user.Email != "john@example.com" {
		t.Fatalf("expected to find user1, found=%v user=%#v", found, user)
	}

	user, found, err = repo.TryFindByID(ctx, "missing")
	if err != nil {
		t.Fatalf("TryFindByID missing failed: %v", err)
	}
	if found || user != nil {
		t.Fatalf("expected missing user to return nil,false; got %#v,%v", user, found)
	}
}

func TestJSONRepository_FindFirstByField(t *testing.T) {
	ctx := context.Background()
	client := newMockJSON()
	kv := newMockKV()
	repo := repository.NewJSONRepository[TestUser](client, kv, "user")

	user := &TestUser{ID: "user1", Name: "John Doe", Email: "john@example.com", Age: 30}
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := repo.FindFirstByField(ctx, "Email", "john@example.com")
	if err != nil {
		t.Fatalf("FindFirstByField failed: %v", err)
	}
	if got.ID != "user1" {
		t.Fatalf("expected user1, got %s", got.ID)
	}

	_, err = repo.FindFirstByField(ctx, "Email", "missing@example.com")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestJSONRepository_CountAndFindFirstIgnoreIndexKeys(t *testing.T) {
	ctx := context.Background()
	client := newMockJSON()
	kv := newMockKV()
	repo := repository.NewJSONRepository[TestUser](client, kv, "user")

	kv.data["user:user1"] = []byte("exists")
	kv.data["user:idx:email:john@example.com:user1"] = []byte("1")
	payload, err := json.Marshal(&TestUser{ID: "user1", Name: "John Doe", Email: "john@example.com", Age: 30})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	client.data["user:user1"] = payload

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected index keys to be ignored, got count %d", count)
	}

	got, err := repo.FindFirst(ctx)
	if err != nil {
		t.Fatalf("FindFirst failed: %v", err)
	}
	if got.ID != "user1" {
		t.Fatalf("expected user1, got %s", got.ID)
	}
}

func TestJSONRepository_FindFirstByFields(t *testing.T) {
	ctx := context.Background()
	client := newMockJSON()
	kv := newMockKV()
	repo := repository.NewJSONRepository[TestUser](client, kv, "user")

	users := []*TestUser{
		{ID: "user1", Name: "John Doe", Email: "john@example.com", Age: 30},
		{ID: "user2", Name: "Jane Doe", Email: "jane@example.com", Age: 30},
	}
	for _, user := range users {
		if err := repo.Save(ctx, user); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	got, err := repo.FindFirstByFields(ctx, map[string]string{
		"Email": "jane@example.com",
		"Age":   "30",
	})
	if err != nil {
		t.Fatalf("FindFirstByFields failed: %v", err)
	}
	if got.ID != "user2" {
		t.Fatalf("expected user2, got %s", got.ID)
	}
}
