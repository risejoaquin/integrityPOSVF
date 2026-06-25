package model

import "encoding/json"

type CustomFieldDefinition struct {
	ID           string          `json:"id"`
	Entity       string          `json:"entity"`
	FieldName    string          `json:"field_name"`
	FieldType    string          `json:"field_type"`
	Options      json.RawMessage `json:"options,omitempty"`
	DisplayOrder int             `json:"display_order"`
}
