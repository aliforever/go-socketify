package socketify

import "github.com/gorilla/websocket"

type Client struct {
	ws *websocket.Conn
}

func newClient(ws *websocket.Conn) (c *Client) {
	c = &Client{ws: ws}
	return
}

func (c *Client) Updates() {

}
