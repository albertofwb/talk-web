package model

import (
	"time"
)

type Message struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id" gorm:"not null;index"`
	Username   string     `json:"username" gorm:"not null"`
	Text       string     `json:"text" gorm:"not null"` // 用户说的话（STT识别结果）
	Reply      string     `json:"reply"`                 // AI回复的内容
	ReplyAudio string     `json:"reply_audio"`           // TTS 生成的音频文件 URL
	Status     string     `json:"status" gorm:"not null;default:'sent'"` // sent, replied, timeout
	SentAt     time.Time  `json:"sent_at" gorm:"not null"`
	RepliedAt  *time.Time `json:"replied_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
