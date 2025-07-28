package models

import (
	"encoding/json"
	"iter"
	"time"

	"github.com/rs/zerolog"
)

const (
	dataFieldName      = "data"
	schemaFieldName    = "schemaURI"
	originFieldName    = "origin"
	timestampFieldName = "timestamp"
	defaultTSFormat    = time.RFC3339
)

// Message is a data object for the stream buffer
// The idea with it is to avoid marshalling/unmarshalling as it can hurt performance
// Must be maintained with care to avoid assigning wrong or switched field values
type Message struct {
	Origin    string          `json:"origin" validate:"required"`
	SchemaURI string          `json:"schemaURI" validate:"required"`
	Timestamp time.Time       `json:"timestamp" validate:"required"`
	Data      json.RawMessage `json:"data" validate:"required,json"`
}

func (m *Message) MarshalZerologObject(e *zerolog.Event) {
	e.Str(originFieldName, m.Origin).
		Str(schemaFieldName, m.SchemaURI).
		Time(timestampFieldName, m.Timestamp).
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

	ts := f(timestampFieldName)
	m.Timestamp, _ = time.ParseInLocation(defaultTSFormat, ts, time.Local)
}

func (m *Message) FromValkeyValue(v map[string]string) {
	m.Data = json.RawMessage(v[dataFieldName])
	m.Origin = v[originFieldName]
	m.SchemaURI = v[schemaFieldName]
	m.Timestamp, _ = time.ParseInLocation(defaultTSFormat, v[timestampFieldName], time.Local)
}

func (m *Message) ToValue() []string {
	return []string{
		originFieldName, m.Origin,
		schemaFieldName, m.SchemaURI,
		timestampFieldName, m.Timestamp.Format(defaultTSFormat),
		dataFieldName, string(m.Data),
	}
}

func (m *Message) Iter() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for i, v := range map[string]string{
			originFieldName:    m.Origin,
			schemaFieldName:    m.SchemaURI,
			timestampFieldName: m.Timestamp.Format(defaultTSFormat),
			dataFieldName:      string(m.Data),
		} {
			if !yield(i, v) {
				return
			}
		}
	}
}
