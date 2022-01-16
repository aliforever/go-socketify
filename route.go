package socketify

import (
	"log"
	"net/http"
)

func (s *Socketify) websocketUpgrade(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	client := newClient(s, c)
	s.clients <- client
}
