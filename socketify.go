package socketify

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type Socketify struct {
	opts    *options
	upgrade *websocket.Upgrader
	server  *http.Server
	clients chan *Client
}

func New(opts *options) (s *Socketify) {
	if opts == nil {
		opts = defaultOptions()
	} else {
		opts.fillDefaults()
	}

	var upgrade = &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	}}

	if opts.checkOrigin != nil {
		upgrade.CheckOrigin = opts.checkOrigin
	}

	s = &Socketify{
		opts:    opts,
		server:  &http.Server{Addr: opts.address, Handler: opts.serveMux},
		upgrade: upgrade,
		clients: make(chan *Client),
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
