package models

import "encoding/json"

const (
	DataFieldName   = "data"
	SchemaFieldName = "schema"
)

type Message struct {
	SchemaURI string          `json:"schemaURI"`
	Data      json.RawMessage `json:"data"`
}
