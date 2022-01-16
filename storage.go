package socketify

import "sync"

type storage struct {
	m       sync.Mutex
	clients map[string]*Client
}

func newStorage() *storage {
	return &storage{
		clients: map[string]*Client{},
	}
}

func (s *storage) addClient(c *Client) {
	s.m.Lock()
	defer s.m.Unlock()

	s.clients[c.id] = c
}

func (s *storage) GetClientByID(clientID string) *Client {
	s.m.Lock()
	defer s.m.Unlock()

	return s.clients[clientID]
}

func (s *storage) removeClientByID(clientID string) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.clients, clientID)
}
