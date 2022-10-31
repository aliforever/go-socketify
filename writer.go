package socketify

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type writer struct {
	ch     chan messageType
	logger Logger
}

func newWriter(ch chan messageType, logger Logger) *writer {
	w := &writer{ch: ch, logger: logger}
	return w
}

func (w *writer) WriteUpdate(updateType string, data interface{}) (err error) {
	jm := newJSONMessage(serverUpdate{
		Type: updateType,
		Data: data,
	})

	w.ch <- jm

	return <-jm.err
}

func (w *writer) WriteRawUpdate(data interface{}) (err error) {
	jm := newJSONMessage(data)

	w.ch <- jm

	return <-jm.err
}

func (w *writer) WriteBinaryBytes(data []byte) (err error) {
	jm := newBinaryMessage(data)

	w.ch <- jm

	return <-jm.err
}

func (w *writer) WriteBinaryText(data []byte) (err error) {
	jm := newBinaryTextMessage(data)

	w.ch <- jm

	return <-jm.err
}

func (w *writer) WriteText(data string) (err error) {
	jm := newTextMessage(data)

	w.ch <- jm

	return <-jm.err
}

func (w *writer) processWriter(ws *websocket.Conn) {
	for update := range w.ch {
		data, err := update.Data()
		if err != nil {
			go func(update messageType, err error) {
				update.Err() <- err
			}(update, err)
			w.logger.Error("Error getting message data", err, fmt.Sprintf("update: %+v . RemoteAddr: %s", update, ws.RemoteAddr().String()))
			continue
		}

		err = ws.WriteMessage(update.Type(), data)
		if err != nil {
			w.logger.Error("Error writing JSON", err, fmt.Sprintf("update: %+v . RemoteAddr: %s", update, ws.RemoteAddr().String()))
		}
		go func(update messageType, err error) {
			update.Err() <- err
		}(update, err)
	}
}

func (c *Connection) WriteInternalUpdate(update []byte) {
	c.internalUpdates <- update
	// TODO: Decide to move this to goroutine or not, because people might forget to do so in their application leading \
	//  to a deadlock
}

func (c *Connection) ping() {
	if c.keepAlive != 0 {
		ticker := time.NewTicker(c.keepAlive)
		for range ticker.C {
			err := c.ws.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}
