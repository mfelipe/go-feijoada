package models

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func TestMessage_ExpectedFields(t *testing.T) {
	msg := Message{}

	msgValue := reflect.ValueOf(msg)
	msgType := msgValue.Type()

	for i := 0; i < msgValue.NumField(); i++ {
		tag := msgType.Field(i).Tag.Get("json")
		switch tag {
		case originFieldName, schemaFieldName, dataFieldName, timestampFieldName:
		default:
			t.Errorf("Unexpected field with tag %s. UPDATE YOUR TESTS!", tag)
		}
	}
}

func TestMessage_FromRedisValue(t *testing.T) {
	testTime := time.Now()
	formattedTime := testTime.Format(defaultTSFormat)

	type args struct {
		v map[string]any
	}
	tests := []struct {
		name        string
		args        args
		expected    Message
		isValid     bool
		expectPanic bool
	}{
		{"All",
			args{map[string]any{
				originFieldName:    "some origin",
				schemaFieldName:    "some schema",
				dataFieldName:      `{"some":"json"}`,
				timestampFieldName: formattedTime,
			}},
			Message{
				Origin:    "some origin",
				SchemaURI: "some schema",
				Data:      json.RawMessage(`{"some":"json"}`),
				Timestamp: testTime.Truncate(time.Second), // Truncate to match RFC3339 format
			},
			true,
			false,
		},
		{"Missing origin",
			args{map[string]any{
				schemaFieldName:    "some schema",
				dataFieldName:      `{"some":"json"}`,
				timestampFieldName: formattedTime,
			}},
			Message{
				SchemaURI: "some schema",
				Data:      json.RawMessage(`{"some":"json"}`),
				Timestamp: testTime.Truncate(time.Second),
			},
			false,
			false,
		},
		{"Missing schemaURI",
			args{map[string]any{
				originFieldName:    "some origin",
				dataFieldName:      `{"some":"json"}`,
				timestampFieldName: formattedTime,
			}},
			Message{
				Origin:    "some origin",
				Data:      json.RawMessage(`{"some":"json"}`),
				Timestamp: testTime.Truncate(time.Second),
			},
			false,
			false,
		},
		{"Missing Data",
			args{map[string]any{
				originFieldName:    "some origin",
				schemaFieldName:    "some schema",
				timestampFieldName: formattedTime,
			}},
			Message{
				Origin:    "some origin",
				SchemaURI: "some schema",
				Data:      json.RawMessage(""),
				Timestamp: testTime.Truncate(time.Second),
			},
			false,
			false,
		},
		{"Missing Timestamp",
			args{map[string]any{
				originFieldName: "some origin",
				schemaFieldName: "some schema",
				dataFieldName:   `{"some":"json"}`,
			}},
			Message{
				Origin:    "some origin",
				SchemaURI: "some schema",
				Data:      json.RawMessage(`{"some":"json"}`),
			},
			false,
			false,
		},
		{"Wrong data type",
			args{map[string]any{
				originFieldName: 2,
			}},
			Message{},
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{}

			if tt.expectPanic {
				assert.Panics(t, func() {
					m.FromRedisValue(tt.args.v)
				})
				return
			}

			m.FromRedisValue(tt.args.v)

			// For test cases without timestamp, we need to copy the auto-generated timestamp
			if _, ok := tt.args.v[timestampFieldName]; !ok {
				tt.expected.Timestamp = m.Timestamp
			}

			assert.Equal(t, tt.expected, *m)

			err := validate.Struct(m)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMessage_FromValkeyValue(t *testing.T) {
	testTime := time.Now()
	formattedTime := testTime.Format(defaultTSFormat)

	type args struct {
		v map[string]string
	}
	tests := []struct {
		name     string
		args     args
		expected Message
		isValid  bool
	}{
		{"All",
			args{map[string]string{
				originFieldName:    "some origin",
				schemaFieldName:    "some schema",
				dataFieldName:      `{"some":"json"}`,
				timestampFieldName: formattedTime,
			}},
			Message{
				Origin:    "some origin",
				SchemaURI: "some schema",
				Data:      json.RawMessage(`{"some":"json"}`),
				Timestamp: testTime.Truncate(time.Second),
			},
			true,
		},
		{"Missing origin",
			args{map[string]string{
				schemaFieldName:    "some schema",
				dataFieldName:      `{"some":"json"}`,
				timestampFieldName: formattedTime,
			}},
			Message{
				SchemaURI: "some schema",
				Data:      json.RawMessage(`{"some":"json"}`),
				Timestamp: testTime.Truncate(time.Second),
			},
			false,
		},
		{"Missing schemaURI",
			args{map[string]string{
				originFieldName:    "some origin",
				dataFieldName:      `{"some":"json"}`,
				timestampFieldName: formattedTime,
			}},
			Message{
				Origin:    "some origin",
				Data:      json.RawMessage(`{"some":"json"}`),
				Timestamp: testTime.Truncate(time.Second),
			},
			false,
		},
		{"Missing Data",
			args{map[string]string{
				originFieldName:    "some origin",
				schemaFieldName:    "some schema",
				timestampFieldName: formattedTime,
			}},
			Message{
				Origin:    "some origin",
				SchemaURI: "some schema",
				Data:      json.RawMessage(""),
				Timestamp: testTime.Truncate(time.Second),
			},
			false,
		},
		{"Missing Timestamp",
			args{map[string]string{
				originFieldName: "some origin",
				schemaFieldName: "some schema",
				dataFieldName:   `{"some":"json"}`,
			}},
			Message{
				Origin:    "some origin",
				SchemaURI: "some schema",
				Data:      json.RawMessage(`{"some":"json"}`),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{}

			m.FromValkeyValue(tt.args.v)

			// For test cases without timestamp, we need to copy the auto-generated timestamp
			if _, ok := tt.args.v[timestampFieldName]; !ok {
				tt.expected.Timestamp = m.Timestamp
			}

			assert.Equal(t, tt.expected, *m)

			err := validate.Struct(m)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMessage_Iter(t *testing.T) {
	testTime := time.Now().Truncate(time.Second)
	msg := &Message{
		Origin:    "test-origin",
		SchemaURI: "test-schema",
		Data:      json.RawMessage(`{"test":"data"}`),
		Timestamp: testTime,
	}

	// Create a map to track which fields we've seen
	seen := make(map[string]string)

	// Use the iterator
	iter := msg.Iter()
	iter(func(key, value string) bool {
		seen[key] = value
		return true
	})

	// Verify all fields are present with correct values
	assert.Equal(t, msg.Origin, seen[originFieldName], "Origin field mismatch")
	assert.Equal(t, msg.SchemaURI, seen[schemaFieldName], "SchemaURI field mismatch")
	assert.Equal(t, string(msg.Data), seen[dataFieldName], "Data field mismatch")
	assert.Equal(t, msg.Timestamp.Format(defaultTSFormat), seen[timestampFieldName], "Timestamp field mismatch")

	// Verify early termination works
	count := 0
	iter(func(key, value string) bool {
		count++
		return false // Stop after first item
	})
	assert.Equal(t, 1, count, "Iterator should stop after first item when returning false")
}

func TestMessage_ToValue(t *testing.T) {
	testTime := time.Now().Truncate(time.Second)
	tests := []struct {
		name     string
		message  Message
		expected []string
	}{
		{
			name: "Complete message",
			message: Message{
				Origin:    "test-origin",
				SchemaURI: "test-schema",
				Data:      json.RawMessage(`{"test":"data"}`),
				Timestamp: testTime,
			},
			expected: []string{
				originFieldName, "test-origin",
				schemaFieldName, "test-schema",
				dataFieldName, `{"test":"data"}`,
				timestampFieldName, testTime.Format(defaultTSFormat),
			},
		},
		{
			name: "Empty fields",
			message: Message{
				Origin:    "",
				SchemaURI: "",
				Data:      json.RawMessage(``),
				Timestamp: time.Time{},
			},
			expected: []string{
				originFieldName, "",
				schemaFieldName, "",
				dataFieldName, ``,
				timestampFieldName, time.Time{}.Format(defaultTSFormat),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.message.ToValue()
			assert.Equal(t, tt.expected, got, "ToValue() returned unexpected slice")

			// Verify the length is always correct (should be 4 fields * 2 for key-value pairs)
			assert.Equal(t, 8, len(got), "ToValue() should always return 8 elements")

			// Verify field names are in the correct positions
			assert.Equal(t, originFieldName, got[0], "First element should be origin field name")
			assert.Equal(t, schemaFieldName, got[2], "Third element should be schema field name")
			assert.Equal(t, dataFieldName, got[4], "Fifth element should be data field name")
			assert.Equal(t, timestampFieldName, got[6], "Seventh element should be timestamp field name")
		})
	}
}
