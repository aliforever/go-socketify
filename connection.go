package socketify

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"sync"
	"time"
)

type Connection struct {
	*writer

	id                  string
	server              *Server
	ws                  *websocket.Conn
	internalUpdates     chan []byte
	handlers            map[string]mapper
	rawHandler          func(message []byte)
	handlersLocker      sync.Mutex
	upgradeRequest      *http.Request
	closed              chan bool
	attributes          map[string]interface{}
	attributesLocker    sync.Mutex
	onClose             func()
	keepAlive           time.Duration
	middleware          func(message []byte) error
	middlewareForUpdate func(updateType string, data json.RawMessage) error
	clientErrors        chan error
	encryptionFields    *encryptionFields
}

func newConnection(server *Server, ws *websocket.Conn, upgradeRequest *http.Request, clientID string, encryptionFields *encryptionFields) (c *Connection) {
	wr := make(chan messageType)

	c = &Connection{
		id:               clientID,
		server:           server,
		ws:               ws,
		writer:           newWriter(wr, server.opts.logger),
		handlers:         map[string]mapper{},
		closed:           make(chan bool),
		upgradeRequest:   upgradeRequest,
		attributes:       map[string]interface{}{},
		internalUpdates:  make(chan []byte),
		clientErrors:     make(chan error),
		encryptionFields: encryptionFields,
	}

	go c.processWriter(ws)

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

func (c *Connection) SetAttribute(key string, val interface{}) {
	c.attributesLocker.Lock()
	defer c.attributesLocker.Unlock()

	c.attributes[key] = val
}

func (c *Connection) GetAttribute(key string) (val interface{}, exists bool) {
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
// For the second argument you should pass your handler inside DataMapper as follows: socketify.DataMapper[T](handler)
// Care: If you use this method for an updateType, you won't receive the respected update in your listener
func (c *Connection) HandleUpdate(updateType string, handler mapper) {
	c.handlersLocker.Lock()
	defer c.handlersLocker.Unlock()
	c.handlers[updateType] = handler
}

func (c *Connection) Server() *Server {
	return c.server
}

func (c *Connection) NextReader() (messageType int, r io.Reader, err error) {
	return c.ws.NextReader()
}

func (c *Connection) NextWriterBinary() (r io.Writer, err error) {
	return c.ws.NextWriter(websocket.BinaryMessage)
}

func (c *Connection) NextWriterText() (r io.Writer, err error) {
	return c.ws.NextWriter(websocket.TextMessage)
}

func (c *Connection) NextWriterCloseMessage() (r io.Writer, err error) {
	return c.ws.NextWriter(websocket.CloseMessage)
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

		if rawHandler := c.getRawHandler(); rawHandler != nil {
			rawHandler(message)
			continue
		}

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
		if handler := c.getHandlerByType(update.Type); handler != nil {
			handler(update.Data)
			continue
		}
	}
}

func (c *Connection) getRawHandler() func(message []byte) {
	c.handlersLocker.Lock()
	defer c.handlersLocker.Unlock()

	if handler := c.rawHandler; handler != nil {
		return handler
	}

	return nil
}

func (c *Connection) getHandlerByType(t string) mapper {
	c.handlersLocker.Lock()
	defer c.handlersLocker.Unlock()

	if handler := c.handlers[t]; handler != nil {
		return handler
	}

	return nil
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
