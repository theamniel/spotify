package socket

import (
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

const (
	HeartbeatWaitTimeout = 5 * time.Second
	HeartbeatTimeout     = 35 * time.Second
	InitializeTimeout    = 15 * time.Second
)

type Socket struct {
	Pool       *SocketPool
	Broadcast  chan *SocketMessage
	Register   chan *SocketClient
	Unregister chan *SocketClient

	mu    sync.RWMutex
	state any
}

func New() *Socket {
	return &Socket{
		Pool: &SocketPool{
			clients: make(map[string]*SocketClient),
		},
		Broadcast:  make(chan *SocketMessage),
		Register:   make(chan *SocketClient),
		Unregister: make(chan *SocketClient),
		state:      nil,
	}
}

func (s *Socket) Handle(conn *websocket.Conn) {
	client := NewClient(conn)
	s.Register <- client
	client.Run()
	s.Unregister <- client
}

func (s *Socket) SetState(val any) {
	s.mu.Lock()
	s.state = val
	s.mu.Unlock()
}

func (s *Socket) GetState() any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *Socket) HasState() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state != nil
}

func (s *Socket) Run() {
	for {
		select {
		case message := <-s.Broadcast:
			for _, client := range s.Pool.GetAll() {
				if client != nil {
					client.Send <- message
				}
			}
		case client := <-s.Register:
			if s.Pool.Has(client.ID) {
				s.Pool.Delete(client.ID)
				client.Close(CloseAlreadyAuthenticated)
			} else {
				client.Send <- &SocketMessage{OP: SocketHello}
				go s.WatchClient(client)
			}
		case client := <-s.Unregister:
			if s.Pool.Has(client.ID) {
				s.Pool.Delete(client.ID)
			}
		}
	}
}

func (s *Socket) WatchClient(client *SocketClient) {
	heartbeat := false
	heartbeatTime := time.NewTicker(InitializeTimeout)
	defer heartbeatTime.Stop()
	for {
		select {
		case message, ok := <-client.Message:
			if ok {
				if message.OP == SocketInitialize {
					if !s.Pool.Has(client.ID) {
						client.Send <- &SocketMessage{SocketDispatch, "INITIAL_STATE", &s.state}
						s.Pool.Set(client.ID, client)
						heartbeatTime.Reset(HeartbeatTimeout)
					} else {
						s.Pool.Delete(client.ID)
						client.Close(CloseAlreadyAuthenticated)
						return
					}
				} else if message.OP == SocketHeartbeat {
					if c, ok := s.Pool.Get(client.ID); ok && c != nil {
						client.Send <- &SocketMessage{OP: SocketHeartbeatACK}
						heartbeatTime.Reset(HeartbeatTimeout)
						heartbeat = false // reset
					} else {
						s.Pool.Delete(client.ID)
						client.Close(CloseNotAuthenticated)
						return
					}
				} else {
					s.Unregister <- client
					client.Close(CloseInvalidOpcode)
					return
				}
			} else {
				s.Unregister <- client
				client.Close(1011) // Close: internal server error
				return
			}
		case <-heartbeatTime.C:
			if s.Pool.Has(client.ID) { // client already register...
				if !heartbeat {
					client.Send <- &SocketMessage{OP: SocketHeartbeat}
					heartbeat = true
					heartbeatTime.Reset(HeartbeatTimeout) // wait 5 sec
					continue
				} else {
					s.Pool.Delete(client.ID)
					client.Close(CloseByServerRequest)
					return
				}
			}
			s.Unregister <- client
			client.Close(CloseByServerRequest)
			return
		}
	}
}
