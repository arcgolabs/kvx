---
title: 'kvx'
linkTitle: 'kvx'
description: 'Strongly Typed Redis / Valkey Object Access and Repository Layer'
weight: 6
---

## Overview

`kvx` is a layered Redis / Valkey access package focused on strongly typed object access, repository-style persistence, and Redis-native capabilities.

## Install

```bash
go get github.com/arcgolabs/kvx@latest
```

## Documentation map

- Minimal repository usage: [Getting Started](./getting-started)
- JSON repository patterns: [JSON repository](./json-repository)
- Real Redis / Valkey adapters: [Adapters (Redis / Valkey)](./adapters)
- Design docs:
    - [Design overview (English)](./overview)
    - [设计说明（中文，完整）](./overview.zh)

## Current capabilities

- Unified `Client` capability interfaces for `KV`, `Hash`, `JSON`, `PubSub`, `Stream`, `Search`, `Script`, and `Lock`
- Metadata-driven mapping based on `kvx` struct tags
- `HashRepository` and `JSONRepository` for strongly typed persistence
- Repository convenience reads for optional lookup and first-match indexed queries
- Secondary-index helper support through repository indexers
- Feature modules for `json`, `pubsub`, `stream`, `search`, and `lock`
- Thin adapters for Redis and Valkey drivers

## Positioning

`kvx` is not trying to be a generic cache abstraction.
It is a Redis / Valkey-oriented object access layer for services that want typed repositories without giving up Redis-native data models.

## Runnable examples (repository)

- In-memory repositories:
    - Hash repository: [examples/hash_repository](https://github.com/arcgolabs/kvx/tree/main/examples/hash_repository)
    - JSON repository: [examples/json_repository](https://github.com/arcgolabs/kvx/tree/main/examples/json_repository)
- Real Redis / Valkey with `testcontainers-go`:
    - Redis adapter: [examples/redis_adapter](https://github.com/arcgolabs/kvx/tree/main/examples/redis_adapter)
    - Redis hash: [examples/redis_hash](https://github.com/arcgolabs/kvx/tree/main/examples/redis_hash)
    - Redis JSON: [examples/redis_json](https://github.com/arcgolabs/kvx/tree/main/examples/redis_json)
    - Redis stream: [examples/redis_stream](https://github.com/arcgolabs/kvx/tree/main/examples/redis_stream)
    - Valkey hash: [examples/valkey_hash](https://github.com/arcgolabs/kvx/tree/main/examples/valkey_hash)
    - Valkey JSON: [examples/valkey_json](https://github.com/arcgolabs/kvx/tree/main/examples/valkey_json)
    - Valkey stream: [examples/valkey_stream](https://github.com/arcgolabs/kvx/tree/main/examples/valkey_stream)
