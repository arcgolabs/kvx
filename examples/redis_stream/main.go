// Package main demonstrates Redis stream operations with kvx.
package main

import (
	"context"
	"fmt"

	"github.com/arcgolabs/kvx"
	redisadapter "github.com/arcgolabs/kvx/adapter/redis"
	"github.com/arcgolabs/kvx/examples/shared"
)

func main() {
	ctx := context.Background()

	container, addr, err := shared.StartContainer(ctx, shared.RedisImage())
	must(err)
	defer func() { must(container.Terminate(ctx)) }()

	adapter, err := redisadapter.New(kvx.ClientOptions{Addrs: []string{addr}})
	must(err)
	defer func() { must(adapter.Close()) }()

	id1, err := adapter.XAdd(ctx, "demo:events", "*", map[string][]byte{
		"type": []byte("user.created"),
		"id":   []byte("u-1"),
	})
	must(err)

	id2, err := adapter.XAdd(ctx, "demo:events", "*", map[string][]byte{
		"type": []byte("user.updated"),
		"id":   []byte("u-1"),
	})
	must(err)

	entries, err := adapter.XRead(ctx, "demo:events", "0-0", 10)
	must(err)

	length, err := adapter.XLen(ctx, "demo:events")
	must(err)

	mustWritef("redis stream addr: %s\n", addr)
	mustWritef("entry ids: %s, %s\n", id1, id2)
	mustWritef("xlen: %d\n", length)
	mustWritef("read entries: %d\n", entries.Len())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustWritef(format string, args ...any) {
	_, err := fmt.Printf(format, args...)
	must(err)
}
