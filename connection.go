package socketify

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type Client struct {
	server   *Socketify
	ws       *websocket.Conn
	updates  chan *Update
	writer   chan *serverUpdate
	handlers map[string]func(message json.RawMessage)
}

func newClient(server *Socketify, ws *websocket.Conn) (c *Client) {
	c = &Client{
		server:   server,
		ws:       ws,
		updates:  make(chan *Update),
		writer:   make(chan *serverUpdate),
		handlers: map[string]func(message json.RawMessage){},
	}
	go c.processWriter()

	return
}

func (c *Client) processWriter() {
	for update := range c.writer {
		c.ws.WriteJSON(serverUpdate{
			Type: update.Type,
			Data: update.Data,
		})
	}
}

func (c *Client) WriteUpdate(updateType string, data interface{}) {
	c.writer <- &serverUpdate{
		Type: updateType,
		Data: data,
	}
}

func (c *Client) ProcessUpdates() (err error) {
	defer c.ws.Close()

	var (
		message []byte
	)
	for {
		_, message, err = c.ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		var update *Update
		jsonErr := json.Unmarshal(message, &update)
		if jsonErr != nil {
			fmt.Println(jsonErr)
			continue
		}

		if update.Type == "" {
			fmt.Println("empty update type")
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
