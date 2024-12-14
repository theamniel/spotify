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
	pool  *Pool[string, *Client]
}

func New[T any]() *Socket[T] {
	return &Socket[T]{
		pool: NewPool[string, *Client](),
	}
}

func (s *Socket[T]) Handle(conn *websocket.Conn) {
	client := NewClient(conn)
	defer s.Unregister(client.ID)

	s.Register(client)
	client.Run()
}

func (s *Socket[T]) SetState(value *T) {
	s.mu.Lock()
	if s.state == nil {
		s.state = value
	} else {
		*s.state = *value
	}
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, client := range s.pool.All() {
		go func(client *Client) { // send to each client in parallel
			if client != nil && client.isConnectionAlive {
				client.Send(msg)
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

func (s *Socket[T]) Unregister(clientID string) {
	if s.pool.Has(clientID) {
		s.pool.Delete(clientID)
	}
}

func (s *Socket[T]) Close() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.pool.Len() > 0 {
		for _, client := range s.pool.All() {
			go func(client *Client) {
				if client != nil {
					client.Close(websocket.CloseNormalClosure, "")
				}
			}(client)
		}
		s.pool.Flush()
	}
}

func (s *Socket[T]) WatchClient(client *Client) {
	heartbeat := false
	heartbeatTime := time.NewTicker(HeartbeatTimeout)
	defer heartbeatTime.Stop()
	defer s.Unregister(client.ID)

	for {
		select {
		case message, ok := <-client.Message:
			if !ok {
				client.Close(websocket.CloseInternalServerErr, "Internal server error")
				return
			}

			switch message.OP {
			case SocketInitialize:
				if !s.pool.Has(client.ID) {
					go client.Send(Dispatch("INITIAL_STATE", &s.state))
					s.pool.Set(client.ID, client)
					continue
				} else {
					client.Close(CloseAlreadyAuthenticated, "Already authenticated") // force disconnect
					return
				}

			case SocketHeartbeat:
				if s.pool.Has(client.ID) {
					go client.Send(HeartbeatACK())
					heartbeatTime.Reset(HeartbeatTimeout)
					if heartbeat {
						heartbeat = false
					} // reset
					continue
				} else {
					client.Close(CloseNotAuthenticated, "Not authenticated")
					return
				}

			default:
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
			// inactive/zombie connection
			client.Close(CloseByServerRequest, "Disconnect by server request")
			return
		}
	}
}
