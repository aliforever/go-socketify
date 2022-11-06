package socketify

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type Client struct {
	*writer

	address string

	ws           *websocket.Conn
	handlersLock sync.Mutex

	handlers map[string]func(json.RawMessage)

	rawHandler func(message []byte)

	onErr func(err error)

	closed chan bool

	onClose func(err error)

	rawMiddleware func(update []byte)
}

func NewClient(address string) *Client {
	ch := make(chan messageType)

	cl := &Client{
		address:  address,
		handlers: map[string]func(json.RawMessage){},
		closed:   make(chan bool),
		writer:   newWriter(ch, logger{}),
	}

	return cl
}

func (c *Client) SetRawHandler(fn func(message []byte)) *Client {
	c.rawHandler = fn

	return c
}

func (c *Client) SetRawMiddleware(fn func(message []byte)) *Client {
	c.rawMiddleware = fn

	return c
}

func (c *Client) SetUpdateTypeHandler(updateType string, fn func(message json.RawMessage)) *Client {
	c.handlersLock.Lock()
	defer c.handlersLock.Unlock()

	c.handlers[updateType] = fn

	return c
}

func (c *Client) SetOnError(fn func(err error)) *Client {
	c.onErr = fn

	return c
}

func (c *Client) SetOnClose(fn func(err error)) *Client {
	c.onClose = fn

	return c
}

func (c *Client) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.address, nil)
	if err != nil {
		return err
	}

	c.ws = conn

	go c.writer.processWriter(conn)

	go c.processUpdates()

	<-c.closed

	return fmt.Errorf("connection_closed")
}

func (c *Client) Close() {
	c.close(fmt.Errorf("manually_called_close"))
}

func (c *Client) close(err error) {
	defer c.ws.Close()

	c.closed <- true

	if c.onClose != nil {
		c.onClose(err)
	}
}

func (c *Client) handlerErr(err error) {
	if c.onErr != nil {
		c.onErr(err)
	}
}

func (c *Client) processUpdates() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			go c.handlerErr(err)
			go c.close(err)
			return
		}

		if c.rawMiddleware != nil {
			go c.rawMiddleware(message)
		}

		if c.rawHandler != nil {
			go c.rawHandler(message)
			continue
		}

		var u *Update

		err = json.Unmarshal(message, &u)
		if err != nil {
			go c.handlerErr(err)
			continue
		}

		c.handlersLock.Lock()
		if handler, ok := c.handlers[u.Type]; ok {
			go handler(u.Data)
			continue
		}
		c.handlersLock.Unlock()
	}
}
