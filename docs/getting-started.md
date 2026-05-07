---
title: 'kvx Getting Started'
linkTitle: 'getting-started'
description: 'Build a typed HashRepository with an in-memory backend and indexes'
weight: 2
---

## Getting started (HashRepository)

This page shows a minimal, runnable in-memory `HashRepository` example with:

- `kvx` struct tags for mapping and indexing
- `repository.NewPreset` for shared repository options
- `FindByID`, `TryFindByID`, `FindFirstByField`, `FindByField`, and `Count`

If you need a real Redis / Valkey connection, see [Adapters (Redis / Valkey)](./adapters).

## Example

```go
package main

import (
	"context"
	"fmt"

	"github.com/arcgolabs/kvx/examples/shared"
	"github.com/arcgolabs/kvx/mapping"
	"github.com/arcgolabs/kvx/repository"
)

func main() {
	ctx := context.Background()
	backend := shared.NewHashBackend()

	preset := repository.NewPreset[shared.User](
		repository.WithKeyBuilder[shared.User](mapping.NewKeyBuilder("demo:user")),
	)

	repo := repository.NewHashRepository[shared.User](backend, backend, "user", preset.HashOptions(
		repository.WithHashCodec[shared.User](mapping.NewHashCodec(nil)),
	)...)

	must(repo.Save(ctx, &shared.User{ID: "u-1", Name: "Alice", Email: "alice@example.com"}))
	must(repo.Save(ctx, &shared.User{ID: "u-2", Name: "Bob", Email: "bob@example.com"}))

	entity, err := repo.FindByID(ctx, "u-1")
	must(err)

	optional, found, err := repo.TryFindByID(ctx, "missing")
	must(err)
	_ = optional

	firstMatch, err := repo.FindFirstByField(ctx, "email", "alice@example.com")
	must(err)

	matches, err := repo.FindByField(ctx, "email", "alice@example.com")
	must(err)

	count, err := repo.Count(ctx)
	must(err)

	fmt.Printf("loaded: %s (%s)\n", entity.Name, entity.Email)
	fmt.Printf("missing found: %v\n", found)
	fmt.Printf("first indexed match: %s\n", firstMatch.ID)
	fmt.Printf("indexed matches: %d\n", matches.Len())
	fmt.Printf("count: %d\n", count)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
```

## Runnable example (repository)

- [examples/hash_repository](https://github.com/arcgolabs/kvx/tree/main/examples/hash_repository)
