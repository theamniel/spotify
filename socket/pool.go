package socket

import "sync"

type Pool struct {
	mu    sync.RWMutex
	items map[string]*Client
}

func NewPool() *Pool {
	return &Pool{items: make(map[string]*Client)}
}

func (s *Pool) Set(k string, v *Client) {
	s.mu.Lock()
	s.items[k] = v
	s.mu.Unlock()
}

func (s *Pool) Get(k string) (*Client, bool) {
	s.mu.RLock()
	c, found := s.items[k]
	s.mu.RUnlock()
	return c, found
}

func (s *Pool) Len() int {
	s.mu.RLock()
	n := len(s.items)
	s.mu.RUnlock()
	return n
}

func (s *Pool) GetAll() map[string]*Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conns := make(map[string]*Client, len(s.items))

	for k, v := range s.items {
		conns[k] = v
	}
	return conns
}

func (s *Pool) Has(k string) bool {
	s.mu.RLock()
	_, ok := s.items[k]
	s.mu.RUnlock()
	return ok
}

func (s *Pool) Delete(k string) {
	s.mu.Lock()
	delete(s.items, k)
	s.mu.Unlock()
}

func (s *Pool) Flush() {
	s.mu.Lock()
	s.items = make(map[string]*Client)
	s.mu.Unlock()
}
