// Package kvx provides a layered key-value abstraction with Redis and Valkey support.
//
// Architecture:
//   - Layer 1: Core client abstraction (KV, Hash, PubSub, Stream, and related interfaces)
//   - Layer 2: Mapping layer (tag parser, metadata, serializers)
//   - Layer 3: Repository layer (HashRepository, JSONRepository, and helpers)
//   - Layer 4: Feature modules (pubsub, stream, json, search, script)
//   - Layer 5: Adapters (redis, valkey)
package kvx
