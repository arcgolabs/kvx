---
title: 'kvx Adapters (Redis / Valkey)'
linkTitle: 'adapters'
description: 'Use kvx adapters with real Redis/Valkey via testcontainers-go'
weight: 4
---

## Adapters

`kvx` provides thin adapters that expose the same capability surface over different drivers:

- `github.com/DaiYuANg/arcgo/kvx/adapter/redis`
- `github.com/DaiYuANg/arcgo/kvx/adapter/valkey`

This page shows a minimal runnable example using `testcontainers-go` to start a real Redis instance and then using `HashRepository` on top of it.

## Example (Redis adapter + testcontainers)

```go
package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/DaiYuANg/arcgo/kvx"
	redisadapter "github.com/DaiYuANg/arcgo/kvx/adapter/redis"
	"github.com/DaiYuANg/arcgo/kvx/repository"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type User struct {
	ID    string `kvx:"id"`
	Name  string `kvx:"name"`
	Email string `kvx:"email,index=email"`
}

func main() {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	must(err)
	defer func() { _ = container.Terminate(ctx) }()

	host, err := container.Host(ctx)
	must(err)

	port, err := container.MappedPort(ctx, "6379/tcp")
	must(err)

	adapter, err := redisadapter.New(kvx.ClientOptions{
		Addrs: []string{net.JoinHostPort(host, port.Port())},
	})
	must(err)
	defer func() { _ = adapter.Close() }()

	repo := repository.NewHashRepository[User](
		adapter,
		adapter,
		"demo:user",
		repository.WithPipeline[User](adapter),
	)

	must(repo.Save(ctx, &User{ID: "u-1", Name: "Alice", Email: "alice@example.com"}))

	entity, err := repo.FindByID(ctx, "u-1")
	must(err)

	fmt.Printf("redis addr: %s\n", net.JoinHostPort(host, port.Port()))
	fmt.Printf("loaded: %s (%s)\n", entity.Name, entity.Email)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
```

## Runnable examples (repository)

- Redis adapter: [examples/kvx/redis_adapter](https://github.com/DaiYuANg/arcgo/tree/main/examples/kvx/redis_adapter)
- Redis hash: [examples/kvx/redis_hash](https://github.com/DaiYuANg/arcgo/tree/main/examples/kvx/redis_hash)
- Redis JSON: [examples/kvx/redis_json](https://github.com/DaiYuANg/arcgo/tree/main/examples/kvx/redis_json)
- Redis stream: [examples/kvx/redis_stream](https://github.com/DaiYuANg/arcgo/tree/main/examples/kvx/redis_stream)
- Valkey hash: [examples/kvx/valkey_hash](https://github.com/DaiYuANg/arcgo/tree/main/examples/kvx/valkey_hash)
- Valkey JSON: [examples/kvx/valkey_json](https://github.com/DaiYuANg/arcgo/tree/main/examples/kvx/valkey_json)
- Valkey stream: [examples/kvx/valkey_stream](https://github.com/DaiYuANg/arcgo/tree/main/examples/kvx/valkey_stream)
