package socketify

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrade = websocket.Upgrader{}

func (s *Socketify) websocketUpgrade(w http.ResponseWriter, r *http.Request) {
	c, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func websocketUpgrade(w http.ResponseWriter, r *http.Request) {

}
