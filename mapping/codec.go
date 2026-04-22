package mapping

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/samber/lo"
)

// Serializer defines the interface for serializing/deserializing values.
type Serializer interface {
	// Marshal serializes a value to bytes.
	Marshal(v any) ([]byte, error)
	// Unmarshal deserializes bytes to a value.
	Unmarshal(data []byte, v any) error
}

// JSONSerializer implements Serializer using encoding/json.
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSONSerializer.
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Marshal implements Serializer.Marshal.
func (s *JSONSerializer) Marshal(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON: %w", err)
	}
	return data, nil
}

// Unmarshal implements Serializer.Unmarshal.
func (s *JSONSerializer) Unmarshal(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}
	return nil
}

// HashCodec encodes/decodes struct fields to/from hash field-value pairs.
type HashCodec struct {
	serializer Serializer
}

// NewHashCodec creates a new HashCodec.
func NewHashCodec(serializer Serializer) *HashCodec {
	if serializer == nil {
		serializer = NewJSONSerializer()
	}

	return &HashCodec{
		serializer: serializer,
	}
}

// Encode encodes an entity to a hash map.
func (c *HashCodec) Encode(entity any, metadata *EntityMetadata) (map[string][]byte, error) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	result, err := lo.ReduceErr(lo.Entries(metadata.Fields), func(result map[string][]byte, entry lo.Entry[string, FieldTag], _ int) (map[string][]byte, error) {
		fieldVal := v.FieldByName(entry.Key)
		if !fieldVal.IsValid() {
			return result, nil
		}

		data, err := c.encodeField(fieldVal)
		if err != nil {
			return nil, fmt.Errorf("encode field %s: %w", entry.Key, err)
		}

		result[storageFieldName(entry.Key, entry.Value)] = data
		return result, nil
	}, make(map[string][]byte, len(metadata.Fields)))
	if err != nil {
		return nil, fmt.Errorf("encode hash fields: %w", err)
	}
	return result, nil
}

// Decode decodes a hash map to an entity.
func (c *HashCodec) Decode(data map[string][]byte, entity any, metadata *EntityMetadata) error {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	storeToField := lo.Associate(lo.Entries(metadata.Fields), func(entry lo.Entry[string, FieldTag]) (string, string) {
		return storageFieldName(entry.Key, entry.Value), entry.Key
	})

	_, err := lo.ReduceErr(lo.Entries(data), func(_ struct{}, entry lo.Entry[string, []byte], _ int) (struct{}, error) {
		fieldName, ok := storeToField[entry.Key]
		if !ok {
			return struct{}{}, nil
		}

		field := v.FieldByName(fieldName)
		if !field.IsValid() || !field.CanSet() {
			return struct{}{}, nil
		}

		if err := c.decodeField(field, entry.Value); err != nil {
			return struct{}{}, fmt.Errorf("decode field %s: %w", fieldName, err)
		}
		return struct{}{}, nil
	}, struct{}{})
	if err != nil {
		return fmt.Errorf("decode hash fields: %w", err)
	}
	return nil
}

// EncodeSingleValue encodes a single value to bytes.
func (c *HashCodec) EncodeSingleValue(value any) ([]byte, error) {
	v := reflect.ValueOf(value)
	return c.encodeField(v)
}

func storageFieldName(fieldName string, fieldTag FieldTag) string {
	if fieldTag.Name != "" {
		return fieldTag.Name
	}
	return fieldName
}

func (c *HashCodec) marshalValue(v any) ([]byte, error) {
	data, err := c.serializer.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal value: %w", err)
	}
	return data, nil
}

func (c *HashCodec) unmarshalValue(data []byte, v any) error {
	if err := c.serializer.Unmarshal(data, v); err != nil {
		return fmt.Errorf("unmarshal value: %w", err)
	}
	return nil
}
