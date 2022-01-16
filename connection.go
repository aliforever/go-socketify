package socketify

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/teris-io/shortid"
)

type Client struct {
	id       string
	server   *Socketify
	ws       *websocket.Conn
	updates  chan *Update
	writer   chan *serverUpdate
	handlers map[string]func(message json.RawMessage)
}

func newClient(server *Socketify, ws *websocket.Conn) (c *Client) {
	c = &Client{
		id:       shortid.MustGenerate(),
		server:   server,
		ws:       ws,
		updates:  make(chan *Update),
		writer:   make(chan *serverUpdate),
		handlers: map[string]func(message json.RawMessage){},
	}
	go c.processWriter()

	return
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) processWriter() {
	for update := range c.writer {
		err := c.ws.WriteJSON(serverUpdate{
			Type: update.Type,
			Data: update.Data,
		})
		if err != nil {
			c.server.opts.logger.Error("Error writing JSON", err, fmt.Sprintf("Update Type: %s - Update Data: %s. RemoteAddr: %s", update.Type, update.Data, c.ws.RemoteAddr().String()))
		}
	}
}

func (c *Client) WriteUpdate(updateType string, data interface{}) {
	c.writer <- &serverUpdate{
		Type: updateType,
		Data: data,
	}
}

func (c *Client) ProcessUpdates() (err error) {
	defer c.close()

	var (
		message []byte
	)
	for {
		_, message, err = c.ws.ReadMessage()
		if err != nil {
			c.server.opts.logger.Error(fmt.Sprintf("Error Reading Message: %s. RemoteAddr: %s", err, c.ws.RemoteAddr().String()))
			return
		}

		var update *Update
		jsonErr := json.Unmarshal(message, &update)
		if jsonErr != nil {
			c.server.opts.logger.Error(fmt.Sprintf("Error Unmarshalling Request: %s. Data: %s. RemoteAddr: %s", jsonErr, message, c.ws.RemoteAddr().String()))
			continue
		}

		if update.Type == "" {
			c.server.opts.logger.Error(fmt.Sprintf("Error Due to Empty Update Type. Data: %s. RemoteAddr: %s", message, c.ws.RemoteAddr().String()))
			continue
		}

		// Check if there's a default handler registered for the updateType and call it
		// If any handlers found, the update will be processed by that handler and won't be passed to the updates channel
		if handler, ok := c.handlers[update.Type]; ok {
			handler(update.Data)
			continue
		}

		c.updates <- update
	}
}

// HandleUpdate registers a default handler for updateType
// Care: If you use this method for an updateType, you won't receive the respected update in your listener
func (c *Client) HandleUpdate(updateType string, handler func(message json.RawMessage)) {
	c.handlers[updateType] = handler
}

func (c *Client) Updates() chan *Update {
	return c.updates
}

func (c *Client) Server() *Socketify {
	return c.server
}

func (c *Client) close() error {
	if c.server.storage != nil {
		c.server.storage.RemoveClientByID(c.id)
	}

	return c.ws.Close()
}
