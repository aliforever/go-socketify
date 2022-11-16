package socketify

import (
	"fmt"
	"github.com/teris-io/shortid"
	"net/http"
)

func (s *Server) websocketUpgrade(w http.ResponseWriter, r *http.Request) {
	var clientID string
	var attributes map[string]interface{}
	var err error

	if s.opts.onConnect != nil {
		clientID, attributes, err = s.opts.onConnect(w, r)
		if err != nil {
			s.opts.logger.Error("Error onConnect", err)
			return
		}
	}

	c, err := s.upgrade.Upgrade(w, r, nil)
	if err != nil {
		s.opts.logger.Error("Error upgrading request", err, fmt.Sprintf("Headers: %+v", r.Header))
		return
	}

	if clientID == "" {
		clientID = shortid.MustGenerate()
	}

	client := newConnection(s, c, r, clientID)
	if s.storage != nil {
		s.storage.addClient(client)
	}

	if attributes != nil {
		for key, val := range attributes {
			client.SetAttribute(key, val)
		}
	}

	s.connections <- client
}
