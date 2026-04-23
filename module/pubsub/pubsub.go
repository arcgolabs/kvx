// Package pubsub provides Pub/Sub functionality.
package pubsub

import (
	"context"
	"errors"

	collectionmapping "github.com/arcgolabs/collectionx/mapping"
	"github.com/arcgolabs/kvx"
	"github.com/samber/lo"
	"github.com/samber/oops"
)

// PubSub provides high-level pub/sub operations.
type PubSub struct {
	client        kvx.PubSub
	subscriptions *collectionmapping.ConcurrentMap[string, kvx.Subscription]
}

// NewPubSub creates a new PubSub instance.
func NewPubSub(client kvx.PubSub) *PubSub {
	return &PubSub{
		client:        client,
		subscriptions: collectionmapping.NewConcurrentMap[string, kvx.Subscription](),
	}
}

// Publish publishes a message to a channel.
func (p *PubSub) Publish(ctx context.Context, channel string, message []byte) error {
	if p == nil || p.client == nil {
		return oops.In("kvx/module/pubsub").
			With("op", "publish", "channel", channel, "message_size", len(message)).
			New("pubsub client is nil")
	}
	if err := p.client.Publish(ctx, channel, message); err != nil {
		return oops.In("kvx/module/pubsub").
			With("op", "publish", "channel", channel, "message_size", len(message)).
			Wrapf(err, "publish message")
	}
	return nil
}

// PublishString publishes a string message to a channel.
func (p *PubSub) PublishString(ctx context.Context, channel, message string) error {
	return p.Publish(ctx, channel, []byte(message))
}

// Subscribe subscribes to a channel.
func (p *PubSub) Subscribe(ctx context.Context, channel string) (<-chan []byte, error) {
	if p == nil || p.client == nil {
		return nil, oops.In("kvx/module/pubsub").
			With("op", "subscribe", "channel", channel).
			New("pubsub client is nil")
	}
	if sub, ok := p.subscriptions.Get(channel); ok {
		return sub.Channel(), nil
	}

	sub, err := p.client.Subscribe(ctx, channel)
	if err != nil {
		return nil, oops.In("kvx/module/pubsub").
			With("op", "subscribe", "channel", channel).
			Wrapf(err, "subscribe to channel")
	}

	actual, loaded := p.subscriptions.GetOrStore(channel, sub)
	if loaded {
		if err := sub.Close(); err != nil {
			return nil, oops.In("kvx/module/pubsub").
				With("op", "subscribe", "channel", channel, "stage", "close_duplicate").
				Wrapf(err, "close duplicate subscription")
		}
		return actual.Channel(), nil
	}

	return sub.Channel(), nil
}

// Unsubscribe unsubscribes from a channel.
func (p *PubSub) Unsubscribe(_ context.Context, channel string) error {
	if p == nil || p.subscriptions == nil {
		return oops.In("kvx/module/pubsub").
			With("op", "unsubscribe", "channel", channel).
			New("pubsub registry is nil")
	}
	if sub, ok := p.subscriptions.LoadAndDelete(channel); ok {
		if err := sub.Close(); err != nil {
			return oops.In("kvx/module/pubsub").
				With("op", "unsubscribe", "channel", channel).
				Wrapf(err, "unsubscribe from channel")
		}
	}

	return nil
}

// Close closes all subscriptions.
func (p *PubSub) Close() error {
	if p == nil || p.subscriptions == nil {
		return oops.In("kvx/module/pubsub").
			With("op", "close").
			New("pubsub registry is nil")
	}
	errs := lo.FilterMap(p.subscriptions.Keys(), func(channel string, _ int) (error, bool) {
		sub, ok := p.subscriptions.LoadAndDelete(channel)
		if !ok {
			return nil, false
		}
		err := sub.Close()
		if err == nil {
			return nil, false
		}
		return oops.In("kvx/module/pubsub").
			With("op", "close", "channel", channel).
			Wrapf(err, "close subscription"), true
	})

	return errors.Join(errs...)
}
