// Package valkey_hash demonstrates Valkey hash operations with kvx.
package main

import (
	"context"
	"fmt"

	"github.com/arcgolabs/kvx"
	valkeyadapter "github.com/arcgolabs/kvx/adapter/valkey"
	"github.com/arcgolabs/kvx/examples/shared"
)

func main() {
	ctx := context.Background()

	container, addr, err := shared.StartContainer(ctx, shared.ValkeyImage())
	must(err)
	defer func() { must(container.Terminate(ctx)) }()

	adapter, err := valkeyadapter.New(kvx.ClientOptions{Addrs: []string{addr}})
	must(err)
	defer func() { must(adapter.Close()) }()

	must(adapter.HSet(ctx, "demo:user:u-1", map[string][]byte{
		"id":    []byte("u-1"),
		"name":  []byte("Alice"),
		"email": []byte("alice@example.com"),
	}))

	name, err := adapter.HGet(ctx, "demo:user:u-1", "name")
	must(err)

	fields, err := adapter.HGetAll(ctx, "demo:user:u-1")
	must(err)

	length, err := adapter.HLen(ctx, "demo:user:u-1")
	must(err)

	mustWritef("valkey hash addr: %s\n", addr)
	mustWritef("name: %s\n", string(name))
	mustWritef("fields: %d\n", len(fields))
	mustWritef("hlen: %d\n", length)
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
