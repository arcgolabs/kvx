package stream

import (
	collectionlist "github.com/arcgolabs/collectionx/list"
	collectionmapping "github.com/arcgolabs/collectionx/mapping"
	"github.com/arcgolabs/kvx"
	"github.com/samber/lo"
	"github.com/samber/oops"
)

func buildByteValues(values map[string]any) (map[string][]byte, error) {
	byteValues, err := lo.ReduceErr(lo.Entries(values), func(byteValues map[string][]byte, entry lo.Entry[string, any], _ int) (map[string][]byte, error) {
		data, err := convertToBytes(entry.Value)
		if err != nil {
			return nil, err
		}
		byteValues[entry.Key] = data
		return byteValues, nil
	}, make(map[string][]byte, len(values)))
	if err != nil {
		return nil, oops.In("kvx/module/stream").
			With("op", "build_stream_values", "field_count", len(values)).
			Wrapf(err, "build stream values")
	}
	return byteValues, nil
}

func convertToBytes(v any) ([]byte, error) {
	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	case nil:
		return []byte(""), nil
	default:
		return marshalJSON(v, "marshal stream value")
	}
}

func limitEntries(entries *collectionlist.List[kvx.StreamEntry], count int64) *collectionlist.List[kvx.StreamEntry] {
	if entries == nil || count <= 0 || count >= int64(entries.Len()) {
		return entries
	}

	return entries.Take(int(count))
}

func streamEntriesFromMultiMap(results *collectionmapping.MultiMap[string, kvx.StreamEntry], streamKey string) *collectionlist.List[kvx.StreamEntry] {
	entries := results.Get(streamKey)
	if len(entries) == 0 {
		return collectionlist.NewList[kvx.StreamEntry]()
	}
	return collectionlist.NewListWithCapacity(len(entries), entries...)
}
