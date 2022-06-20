package socketify

import "encoding/json"

type Update struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type serverUpdate struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
	err  chan error
}
