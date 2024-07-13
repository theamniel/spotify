package socket

import "sync"

type Pool[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

func NewPool[K comparable, V any]() *Pool[K, V] {
	return &Pool[K, V]{items: make(map[K]V)}
}

func (s *Pool[K, V]) Set(k K, v V) {
	s.mu.Lock()
	s.items[k] = v
	s.mu.Unlock()
}

func (s *Pool[K, V]) Get(k K) (V, bool) {
	s.mu.RLock()
	c, found := s.items[k]
	s.mu.RUnlock()
	return c, found
}

func (s *Pool[K, V]) Len() int {
	s.mu.RLock()
	n := len(s.items)
	s.mu.RUnlock()
	return n
}

func (s *Pool[K, V]) All() map[K]V {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conns := make(map[K]V, len(s.items))

	for k, v := range s.items {
		conns[k] = v
	}
	return conns
}

func (s *Pool[K, V]) Has(k K) bool {
	s.mu.RLock()
	_, ok := s.items[k]
	s.mu.RUnlock()
	return ok
}

func (s *Pool[K, V]) Delete(k K) {
	s.mu.Lock()
	delete(s.items, k)
	s.mu.Unlock()
}

func (s *Pool[K, V]) Flush() {
	s.mu.Lock()
	s.items = make(map[K]V)
	s.mu.Unlock()
}
