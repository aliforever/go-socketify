package socketify

import (
	"encoding/json"
	"net/http"
	"sync"
)

type Socketify struct {
	m        sync.Mutex
	opts     *options
	server   *http.Server
	clients  chan *Client
	handlers map[string]func(message json.RawMessage) error
}

func New(opts *options) (s *Socketify) {
	if opts == nil {
		opts = defaultOptions()
	}

	s = &Socketify{
		opts:     opts,
		server:   &http.Server{Addr: opts.address, Handler: opts.serveMux},
		handlers: map[string]func(message json.RawMessage) error{},
	}
	return
}

func (s *Socketify) HandleUpdate(eventName string, handler func(message json.RawMessage) error) {
	s.m.Lock()
	defer s.m.Unlock()
	s.handlers[eventName] = handler
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
