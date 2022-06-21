package socketify

import "github.com/gorilla/websocket"

type messageTypeText struct {
	data string
	err  chan error
}

func (m messageTypeText) Type() int {
	return websocket.TextMessage
}

func (m *messageTypeText) Data() ([]byte, error) {
	return []byte(m.data), nil
}

func (m *messageTypeText) Err() chan error {
	return m.err
}

func newTextMessage(data string) *messageTypeText {
	return &messageTypeText{data: data, err: make(chan error)}
}
