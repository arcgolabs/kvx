// Package main lists the available kvx example entrypoints.
package main

import "fmt"

func main() {
	mustPrintln("Available kvx examples:")
	mustPrintln("  go run ./examples/hash_repository")
	mustPrintln("  go run ./examples/json_repository")
	mustPrintln("  go run ./examples/redis_adapter")
	mustPrintln("  go run ./examples/redis_hash")
	mustPrintln("  go run ./examples/redis_json")
	mustPrintln("  go run ./examples/redis_stream")
	mustPrintln("  go run ./examples/valkey_hash")
	mustPrintln("  go run ./examples/valkey_json")
	mustPrintln("  go run ./examples/valkey_stream")
}

func mustPrintln(args ...any) {
	if _, err := fmt.Println(args...); err != nil {
		panic(err)
	}
}
