package socketify

import "github.com/gorilla/websocket"

type messageTypeBinaryText struct {
	data []byte
	err  chan error
}

func (m messageTypeBinaryText) Type() int {
	return websocket.TextMessage
}

func (m *messageTypeBinaryText) Data() ([]byte, error) {
	return m.data, nil
}

func (m *messageTypeBinaryText) Err() chan error {
	return m.err
}

func newBinaryTextMessage(data []byte) *messageTypeBinaryText {
	return &messageTypeBinaryText{data: data, err: make(chan error)}
}
