package socket

import "sync"

type SocketPool struct {
	sync.RWMutex
	clients map[string]*SocketClient
}

func NewPool() *SocketPool {
	return &SocketPool{
		clients: make(map[string]*SocketClient),
	}
}

func (s *SocketPool) Set(id string, client *SocketClient) {
	s.Lock()
	s.clients[id] = client
	s.Unlock()
}

func (s *SocketPool) Get(id string) (*SocketClient, bool) {
	s.RLock()
	c, ok := s.clients[id]
	s.RUnlock()
	return c, ok
}

func (s *SocketPool) Len() int {
	s.RLock()
	n := len(s.clients)
	s.RUnlock()
	return n
}

func (s *SocketPool) GetAll() map[string]*SocketClient {
	s.RLock()
	conns := make(map[string]*SocketClient, 0)
	for id, client := range s.clients {
		conns[id] = client
	}
	s.RUnlock()
	return conns
}

func (s *SocketPool) Has(id string) bool {
	s.RLock()
	_, ok := s.clients[id]
	s.RUnlock()
	return ok
}

func (s *SocketPool) Delete(id string) {
	s.Lock()
	delete(s.clients, id)
	s.Unlock()
}
