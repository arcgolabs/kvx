---
title: 'kvx overview'
linkTitle: 'overview'
description: 'Design goals and architecture for the Redis / Valkey typed access layer'
weight: 5
draft: false
---

## What kvx Is

`github.com/DaiYuANg/arcgo/kvx` is a **Redis / Valkey–centric** framework for strongly typed object access, repository-style persistence, and feature modules (Pub/Sub, Stream, JSON, Search, Lua). It is **not** a generic cache abstraction for Memcached, SQL, or object storage.

## Goals

- Unified capability surface (`KV`, `Hash`, `JSON`, `PubSub`, `Stream`, `Search`, `Script`, `Lock`) with thin Redis and Valkey adapters.
- Metadata-driven mapping via `kvx` struct tags (primary key, fields, TTL, indexes).
- Repository-style APIs (`HashRepository`, `JSONRepository`) similar in spirit to Spring Data Redis Hash, without copying the whole ecosystem.
- Modular feature packages so advanced capabilities stay optional and composable.

## Non-Goals

- No relational ORM features (joins, relation navigation, lazy loading, multi-table transactions).
- No automatic entity scanning, automatic index migrations, or heavy “magic” infrastructure.
- No pretending Redis is a generic key-value backend for non-Redis semantics.

## Layered Architecture

1. **Core client abstraction** — stable interfaces over drivers; KV, Hash, pipeline, pub/sub, streams, scripts, JSON, search.
2. **Object mapping** — tag parsing, key building, codecs (Hash vs JSON).
3. **Repository layer** — `Save`, `FindByID`, `Delete`, `Exists`, batch helpers, index-assisted lookups.
4. **Feature modules** — `json`, `pubsub`, `stream`, `search`, `lock` helpers on top of the core.
5. **Adapters** — `kvx/adapter/redis`, `kvx/adapter/valkey` isolate driver types from application code.

## Query Model

Repositories focus on **primary key**, **indexed field**, and **search** entry points—not a full query DSL. Complex search belongs in the Search module.

## Keys and Errors

- Keys should be produced by repositories / `KeyBuilder`, not ad-hoc string concatenation in business code.
- Normalize errors into a small set: not found, invalid model, missing PK, serialization, backend/adapter errors.

## Full Chinese Design Doc

The original long-form design document (Chinese) was maintained in `kvx/README.md` and is now published here:

- [kvx 设计概述（中文）](./overview.zh)

For runnable programs, see **Examples** on the [kvx](./_index) index page.
