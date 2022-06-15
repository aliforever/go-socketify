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

func (s *storage) GetClientsByAttributeValue(key, value string) []*Client {
	s.m.Lock()
	defer s.m.Unlock()

	var clients []*Client
	for index, client := range s.clients {
		if val, exists := client.GetAttribute(key); exists {
			if val == value {
				clients = append(clients, s.clients[index])
			}
		}
	}

	return clients
}

// SetClientForID Important: Don't use this method if you're not sure what you're doing
// This might replace the client with an existing alive client
func (s *storage) SetClientForID(id string, client *Client) {
	s.m.Lock()
	defer s.m.Unlock()

	s.clients[id] = client
}

func (s *storage) removeClientByID(clientID string) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.clients, clientID)
}
