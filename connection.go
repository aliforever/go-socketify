package socketify

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/teris-io/shortid"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	id                      string
	server                  *Socketify
	ws                      *websocket.Conn
	updates                 chan *Update // TODO: Remove this or the raw update handler
	writer                  chan interface{}
	handlers                map[string]func(json.RawMessage)
	rawHandler              func(message []byte)
	handlersLocker          sync.Mutex
	upgradeRequest          *http.Request
	closed                  chan bool
	attributes              map[string]string
	attributesLocker        sync.Mutex
	onClose                 func()
	keepAlive               time.Duration
	middleware              func() error
	middlewareForUpdateType func(updateType string) error
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
		attributes:     map[string]string{},
	}
	go c.processWriter()

	return
}

func (c *Client) SetKeepAliveDuration(keepAlive time.Duration) {
	c.keepAlive = keepAlive
}

func (c *Client) SetMiddleware(middleware func() error) {
	c.middleware = middleware
}

func (c *Client) SetUpdateTypeMiddleware(middleware func(updateType string) error) {
	c.middlewareForUpdateType = middleware
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) SetOnClose(onClose func()) {
	c.onClose = onClose
}

func (c *Client) SetAttribute(key, val string) {
	c.attributesLocker.Lock()
	defer c.attributesLocker.Unlock()

	c.attributes[key] = val
}

func (c *Client) GetAttribute(key string) (val string, exists bool) {
	c.attributesLocker.Lock()
	defer c.attributesLocker.Unlock()

	val, exists = c.attributes[key]
	return
}

func (c *Client) UpgradeRequest() *http.Request {
	return c.upgradeRequest
}

func (c *Client) handleIncomingUpdates(errChannel chan error) {
	var (
		message []byte
		err     error
	)

	if c.keepAlive > 0 {
		c.ws.SetReadDeadline(time.Now().Add(c.keepAlive))
		c.ws.SetPingHandler(func(d string) error {
			c.ws.SetReadDeadline(time.Now().Add(c.keepAlive))
			return c.ws.WriteMessage(websocket.PongMessage, nil)
		})
		c.ws.SetPongHandler(func(d string) error {
			return c.ws.SetReadDeadline(time.Now().Add(c.keepAlive))
		})
		go c.ping()
	}

	for {
		_, message, err = c.ws.ReadMessage()
		if err != nil {
			c.server.opts.logger.Error(fmt.Sprintf("Error Reading Message: %s. RemoteAddr: %s", err, c.ws.RemoteAddr().String()))
			errChannel <- err
			return
		}

		if c.middleware != nil {
			if err = c.middleware(); err != nil {
				c.server.opts.logger.Error(fmt.Sprintf("Error From Middleware: %s. RemoteAddr: %s", err, c.ws.RemoteAddr().String()))
				continue
			}
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

		if c.middlewareForUpdateType != nil {
			if err := c.middlewareForUpdateType(update.Type); err != nil {
				c.server.opts.logger.Error(fmt.Sprintf("Error From Middleware: %s. RemoteAddr: %s", err, c.ws.RemoteAddr().String()))
				continue
			}
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

func (c *Client) ProcessUpdates() (err error) {
	errChan := make(chan error)

	go c.handleIncomingUpdates(errChan)

	select {
	case err = <-errChan:
		go c.close()
		return err
	case <-c.closed:
		err = errors.New("connection_closed")
		return
	}
}

// HandleRawUpdate registers a default handler for update
// Note: Add a raw handler if you don't want to follow the API convention {"type": "", "data": {}}
func (c *Client) HandleRawUpdate(handler func(message []byte)) {
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

	if c.onClose != nil {
		go c.onClose()
	}

	return c.ws.Close()
}
