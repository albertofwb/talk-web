package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	MessageQueue = "message_queue"
	InboxPrefix  = "inbox:"
	DefaultBot   = "AlbertClaudeBot"
	DefaultUser  = "AlbertClaudeBot" // 使用同一个 inbox
)

// Config Telegram 配置（继承自 tg/th 命令）
type Config struct {
	RedisAddr string // Redis 地址
	RedisDB   int    // Redis DB
	Recipient string // 默认接收者（bot）
	Username  string // 当前用户名（收件箱）
}

type TelegramClient struct {
	redis  *redis.Client
	ctx    context.Context
	config Config
}

type Message struct {
	Text      string `json:"text"`
	Recipient string `json:"recipient,omitempty"`
	Sender    string `json:"sender,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// NewTelegramClient 创建 Telegram 客户端（使用默认配置）
func NewTelegramClient() *TelegramClient {
	return NewTelegramClientWithConfig(Config{
		RedisAddr: "localhost:6379",
		RedisDB:   0,
		Recipient: DefaultBot,
		Username:  DefaultUser,
	})
}

// NewTelegramClientWithConfig 使用自定义配置创建客户端
func NewTelegramClientWithConfig(config Config) *TelegramClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
		DB:   config.RedisDB,
	})

	return &TelegramClient{
		redis:  rdb,
		ctx:    context.Background(),
		config: config,
	}
}

// SendToTelegram 发送消息到 Telegram (通过 Redis 队列)
func (tc *TelegramClient) SendToTelegram(text string, recipient string) error {
	if recipient == "" {
		recipient = tc.config.Recipient
	}

	msg := Message{
		Text:      text,
		Recipient: recipient,
	}

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	// 推送到 Redis 队列
	err = tc.redis.LPush(tc.ctx, MessageQueue, msgJSON).Err()
	if err != nil {
		return fmt.Errorf("push to redis failed: %w", err)
	}

	return nil
}

// GetFromTelegram 从 Telegram 收件箱获取最新消息
func (tc *TelegramClient) GetFromTelegram(username string) (*Message, error) {
	if username == "" {
		username = tc.config.Username
	}

	key := InboxPrefix + username

	// 获取最新的一条消息
	msgJSON, err := tc.redis.LIndex(tc.ctx, key, 0).Result()
	if err == redis.Nil {
		return nil, nil // 没有消息
	}
	if err != nil {
		return nil, fmt.Errorf("get from redis failed: %w", err)
	}

	var msg Message
	err = json.Unmarshal([]byte(msgJSON), &msg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal message failed: %w", err)
	}

	return &msg, nil
}

// PopFromTelegram 从收件箱获取并删除最新消息
func (tc *TelegramClient) PopFromTelegram(username string) (*Message, error) {
	if username == "" {
		username = tc.config.Username
	}

	key := InboxPrefix + username

	// 弹出最新的一条消息
	msgJSON, err := tc.redis.LPop(tc.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // 没有消息
	}
	if err != nil {
		return nil, fmt.Errorf("pop from redis failed: %w", err)
	}

	var msg Message
	err = json.Unmarshal([]byte(msgJSON), &msg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal message failed: %w", err)
	}

	return &msg, nil
}

// WaitForReply 等待回复（轮询模式）
func (tc *TelegramClient) WaitForReply(username string, timeout time.Duration) (*Message, error) {
	if username == "" {
		username = tc.config.Username
	}

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		msg, err := tc.GetFromTelegram(username)
		if err != nil {
			return nil, err
		}

		if msg != nil {
			// 找到消息，弹出它
			return tc.PopFromTelegram(username)
		}

		// 等待一段时间再检查
		time.Sleep(500 * time.Millisecond)
	}

	return nil, fmt.Errorf("timeout waiting for reply")
}

// Close 关闭 Redis 连接
func (tc *TelegramClient) Close() error {
	return tc.redis.Close()
}
