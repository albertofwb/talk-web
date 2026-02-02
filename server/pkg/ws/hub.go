package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
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

// BroadcastToAll 广播消息给所有在线用户
func (h *Hub) BroadcastToAll(msgType string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for userID, client := range h.clients {
		msg := &Message{
			UserID: userID,
			Type:   msgType,
			Data:   data,
		}
		select {
		case client.send <- msg:
		default:
			// 发送失败，跳过
		}
	}
}

// RedisMessage Redis 收件箱消息格式
type RedisMessage struct {
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

// StartRedisListener 启动 Redis 监听器，自动把新消息推送给所有客户端
func (h *Hub) StartRedisListener(redisAddr string) {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	ctx := context.Background()
	inboxKey := "inbox:AlbertVoiceBot"

	fmt.Printf("[WebSocket Hub] 开始监听 Redis: %s\n", inboxKey)

	for {
		// 阻塞式获取消息 (BRPOP)
		result, err := rdb.BRPop(ctx, 5*time.Second, inboxKey).Result()
		if err == redis.Nil {
			// 超时，继续等待
			continue
		}
		if err != nil {
			fmt.Printf("[Redis Error] %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// result[0] 是 key, result[1] 是 value
		msgJSON := result[1]

		var msg RedisMessage
		if err := json.Unmarshal([]byte(msgJSON), &msg); err != nil {
			fmt.Printf("[Redis Error] JSON 解析失败: %v\n", err)
			continue
		}

		fmt.Printf("[Redis] 收到消息: %s\n", msg.Text)

		// 只处理 to-web 开头的消息
		const prefix = "to-web "
		if len(msg.Text) < len(prefix) || msg.Text[:len(prefix)] != prefix {
			fmt.Printf("[Redis] 忽略非 to-web 消息\n")
			continue
		}

		// 去掉 to-web 前缀
		replyText := msg.Text[len(prefix):]
		fmt.Printf("[Redis] 推送给前端: %s\n", replyText)

		// 广播给所有在线用户
		h.BroadcastToAll("reply", map[string]interface{}{
			"reply":     replyText,
			"timestamp": msg.Timestamp,
		})
	}
}
