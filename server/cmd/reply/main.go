package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	InboxKey = "inbox:AlbertVoiceBot"
)

type Message struct {
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: reply <回复内容>")
		fmt.Println("示例: reply \"你好，这是回复\"")
		os.Exit(1)
	}

	text := os.Args[1]

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer rdb.Close()

	ctx := context.Background()

	// 构建消息
	msg := Message{
		Text:      text,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("❌ JSON 序列化失败: %v\n", err)
		os.Exit(1)
	}

	// 推送到 Redis
	err = rdb.LPush(ctx, InboxKey, msgJSON).Err()
	if err != nil {
		fmt.Printf("❌ 推送到 Redis 失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 已发送到 %s: %s\n", InboxKey, text)
}
