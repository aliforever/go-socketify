package socketify

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/teris-io/shortid"
	"net/http"
	"sync"
)

type Client struct {
	id             string
	server         *Socketify
	ws             *websocket.Conn
	updates        chan *Update
	writer         chan interface{}
	handlers       map[string]func(json.RawMessage)
	rawHandler     func(message json.RawMessage)
	handlersLocker sync.Mutex
	upgradeRequest *http.Request
	closed         chan bool
}

func newClient(server *Socketify, ws *websocket.Conn, upgradeRequest *http.Request) (c *Client) {
	c = &Client{
		id:             shortid.MustGenerate(),
		server:         server,
		ws:             ws,
		updates:        make(chan *Update),
		writer:         make(chan interface{}),
		handlers:       map[string]func(message json.RawMessage){},
		closed:         make(chan bool),
		upgradeRequest: upgradeRequest,
	}
	go c.processWriter()

	return
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) processWriter() {
	for update := range c.writer {
		err := c.ws.WriteJSON(update)
		if err != nil {
			c.server.opts.logger.Error("Error writing JSON", err, fmt.Sprintf("update: %+v . RemoteAddr: %s", update, c.ws.RemoteAddr().String()))
		}
	}
}

func (c *Client) UpgradeRequest() *http.Request {
	return c.upgradeRequest
}

func (c *Client) WriteUpdate(updateType string, data interface{}) {
	c.writer <- &serverUpdate{
		Type: updateType,
		Data: data,
	}
}

func (c *Client) WriteRawUpdate(data interface{}) {
	c.writer <- data
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

		c.handlersLocker.Lock()
		if c.rawHandler != nil {
			c.rawHandler(message)
			c.handlersLocker.Unlock()
			continue
		}
		c.handlersLocker.Unlock()

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
		c.handlersLocker.Lock()
		if handler, ok := c.handlers[update.Type]; ok {
			c.handlersLocker.Unlock()
			handler(update.Data)
			continue
		}
		c.handlersLocker.Unlock()

		c.updates <- update
	}
}

// HandleRawUpdate registers a default handler for update
// Note: Add a raw handler if you don't want to follow the API convention {"type": "", "data": {}}
func (c *Client) HandleRawUpdate(handler func(message json.RawMessage)) {
	c.handlersLocker.Lock()
	defer c.handlersLocker.Unlock()
	c.rawHandler = handler
}

// HandleUpdate registers a default handler for updateType
// Care: If you use this method for an updateType, you won't receive the respected update in your listener
func (c *Client) HandleUpdate(updateType string, handler func(message json.RawMessage)) {
	c.handlersLocker.Lock()
	defer c.handlersLocker.Unlock()
	c.handlers[updateType] = handler
}

func (c *Client) Updates() chan *Update {
	return c.updates
}

func (c *Client) Server() *Socketify {
	return c.server
}

func (c *Client) CloseChannel() chan bool {
	return c.closed
}

func (c *Client) Close() error {
	return c.close()
}

func (c *Client) close() error {
	defer func() {
		c.closed <- true
	}()

	if c.server.storage != nil {
		c.server.storage.removeClientByID(c.id)
	}

	return c.ws.Close()
}
