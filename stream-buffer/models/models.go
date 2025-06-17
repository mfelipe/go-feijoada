package models

import (
	"encoding/json"
	"iter"

	"github.com/rs/zerolog"
)

const (
	dataFieldName   = "data"
	schemaFieldName = "schemaURI"
	originFieldName = "origin"
)

// Message is a data object for the stream buffer
// The idea with it is to avoid marshalling/unmarshalling as it can hurt performance
// Must be maintained with care to avoid assigning wrong or switched field values
type Message struct {
	Origin    string          `json:"origin" validate:"required"`
	SchemaURI string          `json:"schemaURI" validate:"required"`
	Data      json.RawMessage `json:"data" validate:"required,json"`
}

func (m *Message) MarshalZerologObject(e *zerolog.Event) {
	e.Str(originFieldName, m.Origin).
		Str(schemaFieldName, m.SchemaURI).
		RawJSON(dataFieldName, m.Data)
}

func (m *Message) FromRedisValue(v map[string]any) {
	f := func(field string) string {
		value, ok := v[field]
		if !ok {
			value = ""
		}
		return value.(string)
	}
	m.Data = json.RawMessage(f(dataFieldName))
	m.Origin = f(originFieldName)
	m.SchemaURI = f(schemaFieldName)
}

func (m *Message) FromValkeyValue(v map[string]string) {
	m.Data = json.RawMessage(v[dataFieldName])
	m.Origin = v[originFieldName]
	m.SchemaURI = v[schemaFieldName]
}

func (m *Message) ToValue() []string {
	return []string{originFieldName, m.Origin, schemaFieldName, m.SchemaURI, dataFieldName, string(m.Data)}
}

func (m *Message) Iter() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for i, v := range map[string]string{
			originFieldName: m.Origin,
			schemaFieldName: m.SchemaURI,
			dataFieldName:   string(m.Data),
		} {
			if !yield(i, v) {
				return
			}
		}
	}
}
