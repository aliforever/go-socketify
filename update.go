package socketify

import "encoding/json"

type Update struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type serverUpdate struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
