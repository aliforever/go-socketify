package socketify

import (
	"github.com/gorilla/websocket"
	"log"
)

type Client struct {
	ws      *websocket.Conn
	updates chan *Update
}

func newClient(ws *websocket.Conn) (c *Client) {
	c = &Client{ws: ws}
	return
}

func (c *Client) Updates() {
	defer c.ws.Close()
	for {
		mt, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.ws.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
