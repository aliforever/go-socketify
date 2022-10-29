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

type Connection struct {
	id                  string
	server              *Server
	ws                  *websocket.Conn
	internalUpdates     chan []byte
	writer              chan messageType
	handlers            map[string]func(json.RawMessage)
	rawHandler          func(message []byte)
	handlersLocker      sync.Mutex
	upgradeRequest      *http.Request
	closed              chan bool
	attributes          map[string]string
	attributesLocker    sync.Mutex
	onClose             func()
	keepAlive           time.Duration
	middleware          func(message []byte) error
	middlewareForUpdate func(updateType string, data json.RawMessage) error
	clientErrors        chan error
}

func newConnection(server *Server, ws *websocket.Conn, upgradeRequest *http.Request) (c *Connection) {
	id := server.opts.idFunc(upgradeRequest)
	if id == "" {
		id = shortid.MustGenerate()
	}

	c = &Connection{
		id:              id,
		server:          server,
		ws:              ws,
		writer:          make(chan messageType),
		handlers:        map[string]func(message json.RawMessage){},
		closed:          make(chan bool),
		upgradeRequest:  upgradeRequest,
		attributes:      map[string]string{},
		internalUpdates: make(chan []byte),
		clientErrors:    make(chan error),
	}

	go c.processWriter()

	return
}

func (c *Connection) SetKeepAliveDuration(keepAlive time.Duration) {
	c.keepAlive = keepAlive
}

func (c *Connection) SetMiddleware(middleware func(message []byte) error) {
	c.middleware = middleware
}

func (c *Connection) SetUpdateTypeMiddleware(middleware func(updateType string, data json.RawMessage) error) {
	c.middlewareForUpdate = middleware
}

func (c *Connection) ID() string {
	return c.id
}

func (c *Connection) SetOnClose(onClose func()) {
	c.onClose = onClose
}

func (c *Connection) SetAttribute(key, val string) {
	c.attributesLocker.Lock()
	defer c.attributesLocker.Unlock()

	c.attributes[key] = val
}

func (c *Connection) GetAttribute(key string) (val string, exists bool) {
	c.attributesLocker.Lock()
	defer c.attributesLocker.Unlock()

	val, exists = c.attributes[key]
	return
}

func (c *Connection) UpgradeRequest() *http.Request {
	return c.upgradeRequest
}

func (c *Connection) Errors() <-chan error {
	return c.clientErrors
}

func (c *Connection) ProcessUpdates() error {
	errChan := make(chan error)

	go c.handleIncomingUpdates(errChan)

	select {
	case err := <-errChan:
		go c.close()
		return err
	case <-c.closed:
		return errors.New("connection_closed")
	}
}

func (c *Connection) InternalUpdates() <-chan []byte {
	return c.internalUpdates
}

// HandleRawUpdate registers a default handler for update
// Note: Add a raw handler if you don't want to follow the API convention {"type": "", "data": {}}
func (c *Connection) HandleRawUpdate(handler func(message []byte)) {
	c.handlersLocker.Lock()
	defer c.handlersLocker.Unlock()
	c.rawHandler = handler
}

// HandleUpdate registers a default handler for updateType
// Care: If you use this method for an updateType, you won't receive the respected update in your listener
func (c *Connection) HandleUpdate(updateType string, handler func(message json.RawMessage)) {
	c.handlersLocker.Lock()
	defer c.handlersLocker.Unlock()
	c.handlers[updateType] = handler
}

func (c *Connection) Server() *Server {
	return c.server
}

func (c *Connection) Close() error {
	return c.close()
}

func (c *Connection) reportError(err error) {
	go func() {
		c.clientErrors <- err
	}()
}

func (c *Connection) handleIncomingUpdates(errChannel chan error) {
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
			c.reportError(err)
			return
		}

		if c.middleware != nil {
			if err = c.middleware(message); err != nil {
				c.server.opts.logger.Error(fmt.Sprintf("Error From Middleware: %s. RemoteAddr: %s", err, c.ws.RemoteAddr().String()))
				c.reportError(err)
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
			c.reportError(jsonErr)
			continue
		}

		if update.Type == "" {
			c.server.opts.logger.Error(fmt.Sprintf("Error Due to Empty Update Type. Data: %s. RemoteAddr: %s", message, c.ws.RemoteAddr().String()))
			c.reportError(errors.New("empty update type"))
			continue
		}

		if c.middlewareForUpdate != nil {
			if err := c.middlewareForUpdate(update.Type, update.Data); err != nil {
				c.server.opts.logger.Error(fmt.Sprintf("Error From Middleware: %s. RemoteAddr: %s", err, c.ws.RemoteAddr().String()))
				c.reportError(err)
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
	}
}

func (c *Connection) close() error {
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
