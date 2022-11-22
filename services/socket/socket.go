package socket

import (
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

const (
	HeartbeatWaitTimeout = 5 * time.Second
	HeartbeatTimeout     = 30 * time.Second
	InitializeTimeout    = 15 * time.Second
)

type Socket struct {
	Pool       *SocketPool
	Broadcast  chan *SocketMessage
	Register   chan *SocketClient
	Unregister chan *SocketClient

	mu     sync.RWMutex
	status interface{}
}

func New() *Socket {
	return &Socket{
		Pool: &SocketPool{
			clients: make(map[string]*SocketClient),
		},
		Broadcast:  make(chan *SocketMessage),
		Register:   make(chan *SocketClient),
		Unregister: make(chan *SocketClient),
		status:     nil,
	}
}

func (s *Socket) Handle(conn *websocket.Conn) {
	client := NewClient(conn)
	s.Register <- client
	client.Run()
	s.Unregister <- client
}

func (s *Socket) SetStatus(val interface{}) {
	s.mu.Lock()
	s.status = val
	s.mu.Unlock()
}

func (s *Socket) GetStatus() interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Socket) HasStatus() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status != nil
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
				client.Close(4005) // Close: Already authenticated.
			} else {
				client.Send <- &SocketMessage{SocketHello, "", &JSON{
					"heartbeat_interval": HeartbeatTimeout / time.Millisecond,
					"session_id":         client.ID,
				}}
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
	abnormalHeartbeat := false
	heartbeat := time.NewTicker(InitializeTimeout) // first time, 10 sec (wait for SocketInitialize Opcode)
	defer heartbeat.Stop()
	for {
		select {
		case message, ok := <-client.Message:
			if ok {
				if message.OP == SocketHeartbeat {
					if c, ok := s.Pool.Get(client.ID); ok && c != nil {
						client.Send <- &SocketMessage{SocketHeartbeatACK, "", nil}
						heartbeat.Reset(HeartbeatTimeout)
					} else {
						s.Pool.Delete(client.ID)
						client.Close(4003) // Close: Not Authenticated.
						return
					}
				} else if message.OP == SocketInitialize {
					if !s.Pool.Has(client.ID) {
						client.Send <- &SocketMessage{SocketDispatch, "INIT_STATE", &s.status}
						s.Pool.Set(client.ID, client)
						heartbeat.Reset(HeartbeatTimeout)
					} else {
						s.Pool.Delete(client.ID)
						client.Close(4005) // Close: Already authenticated
						return
					}
				} else {
					s.Unregister <- client
					client.Close(4001) // Close: Received Invalid Opcode
					return
				}
			} else {
				s.Unregister <- client
				client.Close(1000) // Close: connection close/internal server error
				return
			}
		case <-heartbeat.C:
			if s.Pool.Has(client.ID) { // client already register...
				if !abnormalHeartbeat {
					// client does send heartbeat
					client.Send <- &SocketMessage{SocketHeartbeat, "", nil}
					abnormalHeartbeat = true
					heartbeat.Reset(HeartbeatWaitTimeout) // 5 sec more...
					continue
				} else {
					s.Pool.Delete(client.ID)
					client.Close(1006) //Close: connection reset by peer
					return
				}
			}
			s.Unregister <- client
			client.Close(1006) // Close: connection reset by peer
			return
		}
	}
}
