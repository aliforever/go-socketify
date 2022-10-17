package socketify

import "sync"

// TODO: Some methods should not be exported when we are allowing direct external access to clients

type storage struct {
	m       sync.Mutex
	clients map[string]*Client
}

func newStorage() *storage {
	return &storage{
		clients: map[string]*Client{},
	}
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

	delete(s.clients, client.id)
	client.id = id
	s.clients[id] = client
}

func (s *storage) ClientIDs() (ids []string) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, client := range s.clients {
		ids = append(ids, client.id)
	}

	return ids
}

func (s *storage) addClient(c *Client) {
	s.m.Lock()
	defer s.m.Unlock()

	s.clients[c.id] = c
}

func (s *storage) removeClientByID(clientID string) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.clients, clientID)
}
