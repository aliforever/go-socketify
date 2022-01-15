package socketify

import (
	"net/http"
)

type Socketify struct {
	opts    *options
	server  *http.Server
	clients chan *Client
}

func New(opts *options) (s *Socketify) {
	if opts == nil {
		opts = defaultOptions()
	}

	s = &Socketify{
		opts:   opts,
		server: &http.Server{Addr: opts.address, Handler: opts.serveMux},
	}
	return
}

func (s *Socketify) Listen() (err error) {
	s.opts.serveMux.HandleFunc(s.opts.endpoint, s.websocketUpgrade)
	err = s.server.ListenAndServe()
	return
}

func (s *Socketify) Server() (server *http.Server) {
	return s.server
}

func (s *Socketify) Clients() chan *Client {
	return s.clients
}
