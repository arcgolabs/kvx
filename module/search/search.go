// Package search provides RediSearch functionality.
package search

import (
	"context"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/arcgolabs/kvx"
	"github.com/samber/oops"
)

// Search provides high-level search operations.
type Search struct {
	client kvx.Search
}

// NewSearch creates a new Search instance.
func NewSearch(client kvx.Search) *Search {
	return &Search{client: client}
}

// Index represents a search index.
type Index struct {
	client    kvx.Search
	name      string
	keyPrefix string
	schema    collectionx.List[kvx.SchemaField]
}

// NewIndex creates a new Index instance.
func NewIndex(client kvx.Search, name, keyPrefix string, schema collectionx.List[kvx.SchemaField]) *Index {
	return &Index{
		client:    client,
		name:      name,
		keyPrefix: keyPrefix,
		schema:    schema,
	}
}

// Create creates the search index.
func (i *Index) Create(ctx context.Context) error {
	if i == nil {
		return oops.In("kvx/module/search").
			With("op", "create_index").
			New("index is nil")
	}
	if i.client == nil {
		return oops.In("kvx/module/search").
			With("op", "create_index", "index", i.name, "key_prefix", i.keyPrefix, "schema_field_count", i.schema.Len()).
			New("search client is nil")
	}
	if err := i.client.CreateIndex(ctx, i.name, i.keyPrefix, i.schema.Values()); err != nil {
		return oops.In("kvx/module/search").
			With("op", "create_index", "index", i.name, "key_prefix", i.keyPrefix, "schema_field_count", i.schema.Len()).
			Wrapf(err, "create search index")
	}
	return nil
}

// Drop drops the search index.
func (i *Index) Drop(ctx context.Context) error {
	if i == nil {
		return oops.In("kvx/module/search").
			With("op", "drop_index").
			New("index is nil")
	}
	if i.client == nil {
		return oops.In("kvx/module/search").
			With("op", "drop_index", "index", i.name).
			New("search client is nil")
	}
	if err := i.client.DropIndex(ctx, i.name); err != nil {
		return oops.In("kvx/module/search").
			With("op", "drop_index", "index", i.name).
			Wrapf(err, "drop search index")
	}
	return nil
}

// Search performs a search query on this index.
func (i *Index) Search(ctx context.Context, query string, opts *Options) (*Result, error) {
	if i == nil {
		return nil, oops.In("kvx/module/search").
			With("op", "search_index", "query", query).
			New("index is nil")
	}
	opts = resolveOptions(opts)
	if i.client == nil {
		return nil, oops.In("kvx/module/search").
			With("op", "search_index", "index", i.name, "query", query, "limit", opts.Limit, "sort_by", opts.SortBy, "ascending", opts.Ascending).
			New("search client is nil")
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}

	var keys collectionx.List[string]
	var err error

	if opts.SortBy != "" {
		keys, err = i.client.SearchWithSort(ctx, i.name, query, opts.SortBy, opts.Ascending, limit)
	} else {
		keys, err = i.client.Search(ctx, i.name, query, limit)
	}

	if err != nil {
		return nil, oops.In("kvx/module/search").
			With("op", "search_index", "index", i.name, "query", query, "limit", limit, "sort_by", opts.SortBy, "ascending", opts.Ascending).
			Wrapf(err, "search index")
	}

	return &Result{
		Keys:  keys,
		Total: int64(keys.Len()),
	}, nil
}

// Options contains options for search queries.
type Options struct {
	Limit     int
	SortBy    string
	Ascending bool
}

// DefaultOptions returns default search options.
func DefaultOptions() *Options {
	return &Options{
		Limit:     10,
		Ascending: true,
	}
}

// Result represents the result of a search query.
type Result struct {
	Keys  collectionx.List[string]
	Total int64
}

func resolveOptions(opts *Options) *Options {
	if opts != nil {
		return opts
	}
	return DefaultOptions()
}
