package repository_test

import (
	"context"
	"testing"

	"github.com/arcgolabs/kvx/mapping"
	"github.com/arcgolabs/kvx/repository"
)

type IndexerTestEntity struct {
	ID    string `kvx:"id"`
	Name  string `kvx:"name,index"`
	Email string `kvx:"email,index"`
	Age   int    `kvx:"age,index"`
}

func TestIndexer_IndexEntity(t *testing.T) {
	ctx := context.Background()
	kv := newMockKV()
	indexer := repository.NewIndexer[IndexerTestEntity](kv, "test")

	entity := &IndexerTestEntity{
		ID:    "entity1",
		Name:  "John",
		Email: "john@example.com",
		Age:   30,
	}

	metadata := &mapping.EntityMetadata{
		KeyField: "ID",
		Fields: map[string]mapping.FieldTag{
			"Name":  {Name: "name", Index: true},
			"Email": {Name: "email", Index: true},
			"Age":   {Name: "age", Index: true},
		},
		IndexFields: []string{"Name", "Email", "Age"},
	}

	err := indexer.IndexEntity(ctx, entity, metadata, "test:entity1")
	if err != nil {
		t.Fatalf("IndexEntity failed: %v", err)
	}

	// Verify indexes were created
	expectedKeys := []string{
		"test:idx:name:John:entity1",
		"test:idx:email:john@example.com:entity1",
		"test:idx:age:30:entity1",
	}

	for _, key := range expectedKeys {
		if _, exists := kv.data[key]; !exists {
			t.Errorf("Expected index key %s to exist", key)
		}
	}
}

func TestIndexer_RemoveEntityFromIndexes(t *testing.T) {
	ctx := context.Background()
	kv := newMockKV()
	indexer := repository.NewIndexer[IndexerTestEntity](kv, "test")

	// Pre-populate indexes
	kv.data["test:idx:name:John:entity1"] = []byte("1")
	kv.data["test:idx:email:john@example.com:entity1"] = []byte("1")
	kv.data["test:idx:age:30:entity1"] = []byte("1")

	entity := &IndexerTestEntity{
		ID:    "entity1",
		Name:  "John",
		Email: "john@example.com",
		Age:   30,
	}

	metadata := &mapping.EntityMetadata{
		KeyField: "ID",
		Fields: map[string]mapping.FieldTag{
			"Name":  {Name: "name", Index: true},
			"Email": {Name: "email", Index: true},
			"Age":   {Name: "age", Index: true},
		},
		IndexFields: []string{"Name", "Email", "Age"},
	}

	err := indexer.RemoveEntityFromIndexes(ctx, entity, metadata)
	if err != nil {
		t.Fatalf("RemoveEntityFromIndexes failed: %v", err)
	}

	// Verify indexes were removed
	expectedKeys := []string{
		"test:idx:name:John:entity1",
		"test:idx:email:john@example.com:entity1",
		"test:idx:age:30:entity1",
	}

	for _, key := range expectedKeys {
		if _, exists := kv.data[key]; exists {
			t.Errorf("Expected index key %s to be removed", key)
		}
	}
}

func TestIndexer_GetEntityIDsByField(t *testing.T) {
	ctx := context.Background()
	kv := newMockKV()
	indexer := repository.NewIndexer[IndexerTestEntity](kv, "test")

	// Pre-populate indexes
	kv.data["test:idx:name:John:entity1"] = []byte("1")
	kv.data["test:idx:name:John:entity2"] = []byte("1")
	kv.data["test:idx:name:Jane:entity3"] = []byte("1")

	ids, err := indexer.GetEntityIDsByField(ctx, "name", "John")
	if err != nil {
		t.Fatalf("GetEntityIDsByField failed: %v", err)
	}

	if len(ids) != 2 {
		t.Errorf("Expected 2 IDs, got %d", len(ids))
	}

	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	if !idMap["entity1"] {
		t.Errorf("Expected entity1 in results")
	}
	if !idMap["entity2"] {
		t.Errorf("Expected entity2 in results")
	}
}
