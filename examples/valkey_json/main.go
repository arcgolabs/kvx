// Package valkey_json demonstrates Valkey JSON operations with kvx.
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

	container, addr, err := shared.StartContainer(ctx, shared.ValkeyJSONImage())
	must(err)
	defer func() { must(container.Terminate(ctx)) }()

	adapter, err := valkeyadapter.New(kvx.ClientOptions{Addrs: []string{addr}})
	must(err)
	defer func() { must(adapter.Close()) }()

	must(adapter.JSONSet(ctx, "demo:user:u-1", "$", []byte(`{"id":"u-1","name":"Alice","roles":["admin"]}`), 0))

	document, err := adapter.JSONGet(ctx, "demo:user:u-1", "$")
	must(err)

	must(adapter.JSONSetField(ctx, "demo:user:u-1", "$.name", []byte(`"Alice Smith"`)))

	name, err := adapter.JSONGetField(ctx, "demo:user:u-1", "$.name")
	must(err)

	mustWritef("valkey json addr: %s\n", addr)
	mustWritef("document: %s\n", string(document))
	mustWritef("updated name: %s\n", string(name))
	mustWritef("image: %s\n", shared.ValkeyJSONImage())
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
