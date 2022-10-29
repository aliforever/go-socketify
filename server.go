package socketify

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type Server struct {
	opts    *options
	upgrade *websocket.Upgrader
	server  *http.Server
	clients chan *Connection
	storage *storage
}

func NewServer(opts *options) (s *Server) {
	if opts == nil {
		opts = defaultOptions()
	}

	opts.fillDefaults()

	var upgrade = &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	}}

	if opts.checkOrigin != nil {
		upgrade.CheckOrigin = opts.checkOrigin
	}

	var storage *storage
	if opts.enableStorage {
		storage = newStorage()
	}

	s = &Server{
		opts:    opts,
		server:  &http.Server{Addr: opts.address, Handler: opts.serveMux},
		upgrade: upgrade,
		clients: make(chan *Connection),
		storage: storage,
	}

	return
}

func (s *Server) Listen() (err error) {
	s.opts.serveMux.HandleFunc(s.opts.endpoint, s.websocketUpgrade)
	err = s.server.ListenAndServe()
	return
}

func (s *Server) Server() (server *http.Server) {
	return s.server
}

func (s *Server) Storage() *storage {
	return s.storage
}

func (s *Server) Clients() chan *Connection {
	return s.clients
}
