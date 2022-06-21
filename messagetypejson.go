package socketify

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

type messageTypeJSON struct {
	data any
	err  chan error
}

func (m messageTypeJSON) Type() int {
	return websocket.TextMessage
}

func (m *messageTypeJSON) Data() ([]byte, error) {
	return json.Marshal(m.data)
}

func (m *messageTypeJSON) Err() chan error {
	return m.err
}

func newJSONMessage(data any) *messageTypeJSON {
	return &messageTypeJSON{data: data, err: make(chan error)}
}
