package redis

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/arcgolabs/kvx"
	"github.com/redis/go-redis/v9"
	"github.com/samber/oops"
)

// Adapter implements kvx.Client using go-redis.
type Adapter struct {
	client *redis.Client
}

var _ kvx.Client = (*Adapter)(nil)

// New creates a new Redis adapter.
func New(opts kvx.ClientOptions) (*Adapter, error) {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	kvx.LogDebug(logger, opts.Debug, "kvx redis adapter create start", "addrs", len(opts.Addrs), "db", opts.DB)
	if len(opts.Addrs) == 0 {
		kvx.LogError(logger, "kvx redis adapter create failed", "error", kvx.ErrInvalidClientOptions)
		return nil, oops.In("kvx/adapter/redis").
			With("op", "new", "field", "addrs", "addrs_count", len(opts.Addrs), "db", opts.DB).
			Wrapf(kvx.ErrInvalidClientOptions, "validate redis client options")
	}
	if opts.UseTLS {
		kvx.LogError(logger, "kvx redis adapter create failed", "error", kvx.ErrUnsupportedOption, "reason", "tls")
		return nil, oops.In("kvx/adapter/redis").
			With("op", "new", "field", "use_tls", "addr", opts.Addrs[0], "db", opts.DB).
			Wrapf(kvx.ErrUnsupportedOption, "redis adapter does not support tls yet")
	}
	if opts.MasterName != "" {
		kvx.LogError(logger, "kvx redis adapter create failed", "error", kvx.ErrUnsupportedOption, "reason", "master_name")
		return nil, oops.In("kvx/adapter/redis").
			With("op", "new", "field", "master_name", "addr", opts.Addrs[0], "master_name", opts.MasterName).
			Wrapf(kvx.ErrUnsupportedOption, "redis adapter does not support sentinel master selection")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:            opts.Addrs[0],
		Password:        opts.Password,
		DB:              opts.DB,
		TLSConfig:       nil, // TODO: support TLS
		PoolSize:        opts.PoolSize,
		MinIdleConns:    opts.MinIdleConns,
		ConnMaxLifetime: opts.ConnMaxLifetime,
		ConnMaxIdleTime: opts.ConnMaxIdleTime,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		kvx.LogError(logger, "kvx redis adapter ping failed", "addr", opts.Addrs[0], "error", err)
		return nil, errors.Join(
			oops.In("kvx/adapter/redis").
				With("op", "new", "stage", "ping", "addr", opts.Addrs[0], "db", opts.DB).
				Wrapf(err, "ping redis server"),
			wrapRedisError("close client after failed ping", rdb.Close()),
		)
	}

	kvx.LogDebug(logger, opts.Debug, "kvx redis adapter create done", "addr", opts.Addrs[0])
	return &Adapter{client: rdb}, nil
}

// NewFromClient creates an adapter from an existing redis.Client.
func NewFromClient(client *redis.Client) *Adapter {
	return &Adapter{client: client}
}

// Close closes the client connection.
func (a *Adapter) Close() error {
	return wrapRedisError("close client", a.client.Close())
}
