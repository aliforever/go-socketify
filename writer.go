package socketify

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

func (c *Client) WriteUpdate(updateType string, data interface{}) (err error) {
	jm := newJSONMessage(serverUpdate{
		Type: updateType,
		Data: data,
	})

	c.writer <- jm

	return <-jm.err
}

func (c *Client) WriteRawUpdate(data interface{}) (err error) {
	jm := newJSONMessage(data)

	c.writer <- jm

	return <-jm.err
}

func (c *Client) WriteBinaryBytes(data []byte) (err error) {
	jm := newBinaryMessage(data)

	c.writer <- jm

	return <-jm.err
}

func (c *Client) WriteBinaryText(data []byte) (err error) {
	jm := newBinaryTextMessage(data)

	c.writer <- jm

	return <-jm.err
}

func (c *Client) WriteText(data string) (err error) {
	jm := newTextMessage(data)

	c.writer <- jm

	return <-jm.err
}

func (c *Client) ping() {
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

func (c *Client) processWriter() {
	for update := range c.writer {
		data, err := update.Data()
		if err != nil {
			go func(update messageType, err error) {
				update.Err() <- err
			}(update, err)
			c.server.opts.logger.Error("Error getting message data", err, fmt.Sprintf("update: %+v . RemoteAddr: %s", update, c.ws.RemoteAddr().String()))
			continue
		}

		err = c.ws.WriteMessage(update.Type(), data)
		if err != nil {
			c.server.opts.logger.Error("Error writing JSON", err, fmt.Sprintf("update: %+v . RemoteAddr: %s", update, c.ws.RemoteAddr().String()))
		}
		go func(update messageType, err error) {
			update.Err() <- err
		}(update, err)
	}
}
