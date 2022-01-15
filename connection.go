package socketify

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

type Client struct {
	server  *Socketify
	ws      *websocket.Conn
	updates chan *Update
	writer  chan *serverUpdate
}

func newClient(server *Socketify, ws *websocket.Conn) (c *Client) {
	c = &Client{
		server:  server,
		ws:      ws,
		updates: make(chan *Update),
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
			continue
		}

		if update.Type == "" {
			continue
		}

		c.updates <- update
	}
}

func (c *Client) Updates() chan *Update {
	return c.updates
}
