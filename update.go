package socketify

import "encoding/json"

type Update struct {
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data,omitempty"`
	Extra string          `json:"extra,omitempty"`
}

type serverUpdate struct {
	Type  string      `json:"type"`
	Data  interface{} `json:"data,omitempty"`
	Extra string      `json:"extra,omitempty"`
}
