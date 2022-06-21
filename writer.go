package socketify

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

func (c *Client) WriteUpdate(updateType string, data interface{}) {
	c.writer <- &serverUpdate{
		Type: updateType,
		Data: data,
	}
}

func (c *Client) WriteRawUpdate(data interface{}) {
	c.writer <- data
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
		err := c.ws.WriteJSON(update)
		if err != nil {
			c.server.opts.logger.Error("Error writing JSON", err, fmt.Sprintf("update: %+v . RemoteAddr: %s", update, c.ws.RemoteAddr().String()))
		}
	}
}
