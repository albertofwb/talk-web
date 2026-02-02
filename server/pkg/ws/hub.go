package ws

import (
	"sync"
)

// Hub 管理所有 WebSocket 连接
type Hub struct {
	// 用户ID -> 客户端连接
	clients map[uint]*Client
	mu      sync.RWMutex

	// 注册新客户端
	register chan *Client

	// 注销客户端
	unregister chan *Client

	// 广播消息到指定用户
	broadcast chan *Message
}

// Message WebSocket 消息
type Message struct {
	UserID  uint        `json:"user_id"`
	Type    string      `json:"type"` // reply, status, error
	Data    interface{} `json:"data"`
}

var GlobalHub *Hub

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			client, ok := h.clients[message.UserID]
			h.mu.RUnlock()

			if ok {
				select {
				case client.send <- message:
				default:
					// 发送失败，关闭连接
					h.mu.Lock()
					delete(h.clients, client.UserID)
					close(client.send)
					h.mu.Unlock()
				}
			}
		}
	}
}

// Register 注册客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// SendToUser 发送消息给指定用户
func (h *Hub) SendToUser(userID uint, msgType string, data interface{}) {
	msg := &Message{
		UserID: userID,
		Type:   msgType,
		Data:   data,
	}
	h.broadcast <- msg
}
