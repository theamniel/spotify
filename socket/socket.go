package socket

import (
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

const (
	// Heatbeat if client don't respond
	HeartbeatWaitTimeout = 10 * time.Second
	// Heartbeat
	HeartbeatTimeout = 35 * time.Second
)

type Socket[T any] struct {
	mu    sync.RWMutex
	state *T
	pool  *Pool
}

func New[T any]() *Socket[T] {
	return &Socket[T]{
		pool:  NewPool(),
		state: nil,
	}
}

func (s *Socket[T]) Handle(conn *websocket.Conn) {
	client := NewClient(conn)
	s.Register(client)
	client.Run()
	s.Unregister(client)
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

func (s *Socket[T]) Broadcast(msg *Message) {
	for _, client := range s.pool.GetAll() {
		go func(client *Client) { // send to each client in parallel
			if client != nil && client.IsAlive() {
				go client.Send(msg)
			}
		}(client)
	}
}

func (s *Socket[T]) Register(client *Client) {
	if s.pool.Has(client.ID) {
		s.pool.Delete(client.ID)
		client.Close(CloseAlreadyAuthenticated, "Already authenticated")
		return
	}
	go s.WatchClient(client)
	client.Send(Hello(JSON{"heartbeat_interval": HeartbeatTimeout / time.Millisecond}))
}

func (s *Socket[T]) Unregister(client *Client) {
	if s.pool.Has(client.ID) {
		s.pool.Delete(client.ID)
	}
}

func (s *Socket[T]) WatchClient(client *Client) {
	heartbeat := false
	heartbeatTime := time.NewTicker(HeartbeatTimeout)
	defer heartbeatTime.Stop()

	for {
		select {
		case message, ok := <-client.Message:
			if !ok {
				// s.Unregister <- client
				go s.Unregister(client)
				client.Close(websocket.CloseInternalServerErr, "Internal server error")
				return
			}

			// OPCODE: Initialize (2)
			if message.OP == SocketInitialize {
				if !s.pool.Has(client.ID) {
					go client.Send(Dispatch("INITIAL_STATE", &s.state))
					s.pool.Set(client.ID, client)
					continue
				}
				s.pool.Delete(client.ID)
				client.Close(CloseAlreadyAuthenticated, "Already authenticated") // force disconnect
				return

				// OPCODE: Heartbeat (3)
			} else if message.OP == SocketHeartbeat {
				if s.pool.Has(client.ID) {
					go client.Send(HeartbeatACK())
					heartbeatTime.Reset(HeartbeatTimeout)
					if heartbeat {
						heartbeat = false
					} // reset
					continue
				}
				s.pool.Delete(client.ID)
				client.Close(CloseNotAuthenticated, "Not authenticated")
				return

			} else {
				// s.Unregister <- client
				go s.Unregister(client)
				client.Close(CloseInvalidOpcode, "Invalid opcode")
				return
			}

		case <-heartbeatTime.C:
			if s.pool.Has(client.ID) { // client already register...
				if !heartbeat {
					go client.Send(Heartbeat())
					heartbeat = true
					heartbeatTime.Reset(HeartbeatTimeout) // wait 5 sec
					continue
				}
			}
			// inactive/"zombie" connection
			s.Unregister(client)
			client.Close(CloseByServerRequest, "Disconnect by server request")
			return
		}
	}
}
