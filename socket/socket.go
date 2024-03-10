package socket

import (
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

const (
	HeartbeatWaitTimeout = 5 * time.Second
	HeartbeatTimeout     = 35 * time.Second
	InitializeTimeout    = 15 * time.Second
)

type Socket[T any] struct {
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client

	mu    sync.RWMutex
	state *T
	pool  *Pool
}

func New[T any]() *Socket[T] {
	return &Socket[T]{
		pool:       NewPool(),
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		state:      nil,
	}
}

func (s *Socket[T]) Handle(conn *websocket.Conn) {
	client := NewClient(conn)
	s.Register <- client
	client.Run()
	s.Unregister <- client
}

func (s *Socket[T]) SetState(val *T) {
	s.mu.Lock()
	s.state = val
	s.mu.Unlock()
}

func (s *Socket[T]) GetState() *T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *Socket[T]) HasState() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state != nil
}

func (s *Socket[T]) Listeners() int {
	return s.pool.Len()
}

func (s *Socket[T]) Run() {
	for {
		select {
		case message := <-s.Broadcast:
			for _, client := range s.pool.GetAll() {
				go func(client *Client) { // send to each client in parallel
					if client != nil && client.IsAlive() {
						client.Send <- message
					}
				}(client)
			}
		case client := <-s.Register:
			if s.pool.Has(client.ID) {
				s.pool.Delete(client.ID)
				client.Close(CloseAlreadyAuthenticated)
			} else {
				client.Send <- &Message{OP: SocketHello}
				go s.WatchClient(client)
			}
		case client := <-s.Unregister:
			if s.pool.Has(client.ID) {
				s.pool.Delete(client.ID)
			}
		}
	}
}

func (s *Socket[T]) WatchClient(client *Client) {
	heartbeat := false
	heartbeatTime := time.NewTicker(InitializeTimeout)
	defer heartbeatTime.Stop()

	for {
		select {
		case message, ok := <-client.Message:
			if !ok {
				s.Unregister <- client
				client.Close(websocket.CloseInternalServerErr)
				return
			}

			// OPCODE: Initialize (2)
			if message.OP == SocketInitialize {
				if !s.pool.Has(client.ID) {
					client.Send <- &Message{SocketDispatch, "INITIAL_STATE", &s.state}
					s.pool.Set(client.ID, client)
					heartbeatTime.Reset(HeartbeatTimeout)
				} else {
					s.pool.Delete(client.ID)
					client.Close(CloseAlreadyAuthenticated)
					return
				}
				// OPCODE: Heartbeat (3)
			} else if message.OP == SocketHeartbeat {
				if s.pool.Has(client.ID) {
					client.Send <- &Message{OP: SocketHeartbeatACK}
					heartbeatTime.Reset(HeartbeatTimeout)
					heartbeat = false // reset
				} else {
					s.pool.Delete(client.ID)
					client.Close(CloseNotAuthenticated)
					return
				}

			} else {
				s.Unregister <- client
				client.Close(CloseInvalidOpcode)
				return
			}

		case <-heartbeatTime.C:
			if s.pool.Has(client.ID) { // client already register...
				if !heartbeat {
					client.Send <- &Message{OP: SocketHeartbeat}
					heartbeat = true
					heartbeatTime.Reset(HeartbeatTimeout) // wait 5 sec
					continue
				}
			}
			s.Unregister <- client
			client.Close(CloseByServerRequest)
			return
		}
	}
}
