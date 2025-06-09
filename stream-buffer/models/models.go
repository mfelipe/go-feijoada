package models

import (
	"encoding/json"
	"iter"

	"github.com/rs/zerolog"
)

const (
	dataFieldName   = "data"
	schemaFieldName = "schema"
)

type Message struct {
	SchemaURI string          `json:"schemaURI"`
	Data      json.RawMessage `json:"data"`
}

func (m *Message) MarshalZerologObject(e *zerolog.Event) {
	e.Str("schemaURI", m.SchemaURI).
		RawJSON("data", m.Data)
}

// MessageFromMap is an overkill function to convert a map to a Message using generics.
// I could have used a two line function for each Redis and Valkey but why not complicate things to avoid code duplication?
// TODO: make a Must function to panic on missing fields
func MessageFromMap[T any](m map[string]T) Message {
	var msg = Message{}

	toString := func(value T) string {
		switch s := any(value).(type) {
		case string:
			return s
		default:
			if ss, ok := s.(string); ok {
				return ss
			}
		}
		return ""
	}

	if v, ok := m[schemaFieldName]; ok {
		msg.SchemaURI = toString(v)
	}

	if v, ok := m[dataFieldName]; ok {
		msg.Data = json.RawMessage(toString(v))
	}
	return msg
}

//func (m *Message) FromRedisValue(v map[string]any) {
//	m.SchemaURI = v[schemaFieldName].(string)
//	m.Data = json.RawMessage(v[dataFieldName].(string))
//}
//
//func (m *Message) FromValkeyValue(v map[string]string) {
//	m.SchemaURI = v[schemaFieldName]
//	m.Data = json.RawMessage(v[dataFieldName])
//}

func (m *Message) ToValue() []string {
	return []string{schemaFieldName, m.SchemaURI, dataFieldName, string(m.Data)}
}

func (m *Message) Iter() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for i, v := range map[string]string{
			schemaFieldName: m.SchemaURI,
			dataFieldName:   string(m.Data),
		} {
			if !yield(i, v) {
				return
			}
		}
	}
}
