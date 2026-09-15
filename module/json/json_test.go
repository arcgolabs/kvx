package json_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/arcgolabs/kvx"
	jsonmodule "github.com/arcgolabs/kvx/module/json"
	"github.com/stretchr/testify/require"
)

type testJSONClient struct {
	documents map[string][]byte
	fields    map[string][]byte
}

func newTestJSONClient() *testJSONClient {
	return &testJSONClient{
		documents: make(map[string][]byte),
		fields:    make(map[string][]byte),
	}
}

func (c *testJSONClient) JSONSet(_ context.Context, key, _ string, value []byte, _ time.Duration) error {
	c.documents[key] = value
	return nil
}

func (c *testJSONClient) JSONGet(_ context.Context, key, _ string) ([]byte, error) {
	value, ok := c.documents[key]
	if !ok {
		return nil, kvx.ErrNil
	}
	return value, nil
}

func (c *testJSONClient) JSONSetField(_ context.Context, key, path string, value []byte) error {
	c.fields[key+":"+path] = value
	return nil
}

func (c *testJSONClient) JSONGetField(_ context.Context, key, path string) ([]byte, error) {
	value, ok := c.fields[key+":"+path]
	if !ok {
		return nil, kvx.ErrNil
	}
	return value, nil
}

func (c *testJSONClient) JSONDelete(_ context.Context, key, path string) error {
	if path == "$" {
		delete(c.documents, key)
		return nil
	}
	delete(c.fields, key+":"+path)
	return nil
}

type testDocument struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestJSONGenericReads(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newTestJSONClient()
	client.documents["user:1"] = []byte(`{"id":"1","name":"Alice"}`)
	client.fields["user:1:$.name"] = []byte(`"Alice"`)
	documents := jsonmodule.NewJSON(client)

	user, err := documents.Get[testDocument](ctx, "user:1")
	require.NoError(t, err)
	require.Equal(t, testDocument{ID: "1", Name: "Alice"}, user)

	name, err := documents.GetPath[string](ctx, "user:1", "$.name")
	require.NoError(t, err)
	require.Equal(t, "Alice", name)
}

func TestJSONGenericArrayPop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newTestJSONClient()
	client.fields["queue:$.items"] = []byte(`[1,2,3]`)
	documents := jsonmodule.NewJSON(client)

	last, err := documents.ArrayPop[int](ctx, "queue", "$.items")
	require.NoError(t, err)
	require.Equal(t, 3, last)
	require.JSONEq(t, `[1,2]`, string(client.fields["queue:$.items"]))
}

func TestJSONGenericMultiGet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newTestJSONClient()
	client.documents["user:1"] = []byte(`{"id":"1","name":"Alice"}`)
	client.documents["user:2"] = []byte(`{"id":"2","name":"Bob"}`)
	documents := jsonmodule.NewJSON(client)

	users, err := documents.MultiGet[testDocument](ctx, []string{"user:1", "missing", "user:2"})
	require.NoError(t, err)
	require.Equal(t, map[string]testDocument{
		"user:1": {ID: "1", Name: "Alice"},
		"user:2": {ID: "2", Name: "Bob"},
	}, users)
}

func TestJSONGenericMultiGetReturnsBackendError(t *testing.T) {
	t.Parallel()

	documents := jsonmodule.NewJSON((*failingJSONClient)(nil))
	_, err := documents.MultiGet[testDocument](context.Background(), []string{"user:1"})
	require.Error(t, err)
	require.False(t, errors.Is(err, kvx.ErrNil))
}

type failingJSONClient struct{}

func (*failingJSONClient) JSONSet(context.Context, string, string, []byte, time.Duration) error {
	return errors.New("backend unavailable")
}

func (*failingJSONClient) JSONGet(context.Context, string, string) ([]byte, error) {
	return nil, errors.New("backend unavailable")
}

func (*failingJSONClient) JSONSetField(context.Context, string, string, []byte) error {
	return errors.New("backend unavailable")
}

func (*failingJSONClient) JSONGetField(context.Context, string, string) ([]byte, error) {
	return nil, errors.New("backend unavailable")
}

func (*failingJSONClient) JSONDelete(context.Context, string, string) error {
	return errors.New("backend unavailable")
}
