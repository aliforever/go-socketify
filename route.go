package socketify

import (
	"fmt"
	"net/http"
)

func (s *Socketify) websocketUpgrade(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrade.Upgrade(w, r, nil)
	if err != nil {
		s.opts.logger.Error("Error upgrading request", err, fmt.Sprintf("Headers: %+v", r.Header))
		return
	}

	client := newClient(s, c, r)
	if s.storage != nil {
		s.storage.addClient(client)
	}
	s.clients <- client
}
