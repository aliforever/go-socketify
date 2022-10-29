package socketify

import "sync"

// TODO: Some methods should not be exported when we are allowing direct external access to clients

type storage struct {
	m       sync.Mutex
	clients map[string]*Connection
}

func newStorage() *storage {
	return &storage{
		clients: map[string]*Connection{},
	}
}

func (s *storage) GetClientByID(clientID string) *Connection {
	s.m.Lock()
	defer s.m.Unlock()

	return s.clients[clientID]
}

func (s *storage) GetClientsByAttributeValue(key, value string) []*Connection {
	s.m.Lock()
	defer s.m.Unlock()

	var clients []*Connection
	for index, client := range s.clients {
		if val, exists := client.GetAttribute(key); exists {
			if val == value {
				clients = append(clients, s.clients[index])
			}
		}
	}

	return clients
}

func (s *storage) ClientIDs() (ids []string) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, client := range s.clients {
		ids = append(ids, client.id)
	}

	return ids
}

func (s *storage) addClient(c *Connection) {
	s.m.Lock()
	defer s.m.Unlock()

	s.clients[c.id] = c
}

func (s *storage) removeClientByID(clientID string) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.clients, clientID)
}
