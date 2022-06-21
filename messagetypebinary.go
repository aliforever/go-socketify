package socketify

import "github.com/gorilla/websocket"

type messageTypeBinary struct {
	data []byte
	err  chan error
}

func (m messageTypeBinary) Type() int {
	return websocket.BinaryMessage
}

func (m *messageTypeBinary) Data() ([]byte, error) {
	return m.data, nil
}

func (m *messageTypeBinary) Err() chan error {
	return m.err
}

func newBinaryMessage(data []byte) *messageTypeBinary {
	return &messageTypeBinary{data: data, err: make(chan error)}
}
