package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"talk-web/server/model"
	"talk-web/server/pkg/stt"
	"talk-web/server/pkg/telegram"
	"talk-web/server/pkg/tts"
	"talk-web/server/pkg/ws"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UploadHandler struct {
	stt *stt.STT
	tts *tts.TTS
	tg  *telegram.TelegramClient
	db  *gorm.DB
	hub *ws.Hub
}

func NewUploadHandler(talkServerURL string, db *gorm.DB, hub *ws.Hub) *UploadHandler {
	return &UploadHandler{
		stt: stt.NewSTT(),
		tts: tts.NewTTS(),
		tg:  telegram.NewTelegramClient(),
		db:  db,
		hub: hub,
	}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	// 获取用户信息
	userID := c.GetUint("user_id")
	username := c.GetString("username")

	// 接收音频文件
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到音频文件"})
		return
	}
	defer file.Close()

	// 保存临时文件
	tmpFile := filepath.Join("/tmp", fmt.Sprintf("talk-upload-%d-%d%s",
		userID, time.Now().Unix(), filepath.Ext(header.Filename)))

	out, err := os.Create(tmpFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建临时文件失败"})
		return
	}
	defer os.Remove(tmpFile) // 清理临时文件

	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	out.Close()

	// STT: 语音转文字
	recognizedText, err := h.stt.Transcribe(tmpFile)
	if err != nil {
		// 记录详细错误
		fmt.Printf("[STT Error] File: %s, Error: %v\n", tmpFile, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "语音识别失败",
			"detail": err.Error(),
			"file": tmpFile,
		})
		return
	}

	// 记录成功的识别
	fmt.Printf("[STT Success] File: %s, Text: %s\n", tmpFile, recognizedText)

	// 保存到数据库
	message := model.Message{
		UserID:   userID,
		Username: username,
		Text:     recognizedText,
		Status:   "sent",
		SentAt:   time.Now(),
	}
	if err := h.db.Create(&message).Error; err != nil {
		fmt.Printf("[DB Error] Failed to save message: %v\n", err)
	}

	// 添加 from-web 前缀标识来源
	telegramText := fmt.Sprintf("from-web %s", recognizedText)

	// 发送到 Telegram
	err = h.tg.SendToTelegram(telegramText, telegram.DefaultBot)
	if err != nil {
		fmt.Printf("[Telegram Error] Failed to send: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "发送到 Telegram 失败",
			"detail": err.Error(),
		})
		return
	}

	fmt.Printf("[Telegram] Message sent: %s\n", telegramText)

	// 立即返回识别结果，不等待回复
	c.JSON(http.StatusOK, gin.H{
		"text":     recognizedText,
		"status":   "sent",
		"message":  "消息已发送，等待回复中...",
		"user_id":  userID,
		"username": username,
	})

	// 异步等待回复并生成 TTS
	go func() {
		fmt.Printf("[Telegram] Waiting for reply (async)...\n")

		// 等待 Telegram 回复（60秒超时）
		replyMsg, err := h.tg.WaitForReply(telegram.DefaultUser, 60*time.Second)
		if err != nil {
			fmt.Printf("[Telegram Error] No reply: %v\n", err)
			// 更新状态为超时
			h.db.Model(&message).Updates(map[string]interface{}{
				"status": "timeout",
			})
			return
		}

		replyText := replyMsg.Text
		fmt.Printf("[Telegram Reply] %s\n", replyText)

		// 只处理 to-web 开头的回复
		const prefix = "to-web "
		shouldPushToWeb := len(replyText) >= len(prefix) && replyText[:len(prefix)] == prefix

		// 去掉前缀用于显示和 TTS
		displayText := replyText
		if shouldPushToWeb {
			displayText = replyText[len(prefix):]
			fmt.Printf("[Telegram] 检测到 to-web 消息，将推送到前端\n")
		} else {
			fmt.Printf("[Telegram] 非 to-web 消息，仅更新数据库\n")
		}

		// TTS: 文字转语音（使用去掉前缀的文本）
		replyAudioPath, err := h.tts.Generate(displayText)
		audioURL := ""
		if err != nil {
			fmt.Printf("[TTS Error] Failed to generate: %v\n", err)
		} else {
			audioFilename := filepath.Base(replyAudioPath)
			audioURL = fmt.Sprintf("/api/audio/%s", audioFilename)
			fmt.Printf("[TTS Success] Audio generated: %s\n", audioFilename)
		}

		// 更新数据库记录（保存去掉前缀的文本）
		now := time.Now()
		h.db.Model(&message).Updates(map[string]interface{}{
			"reply":       displayText,
			"reply_audio": audioURL,
			"status":      "replied",
			"replied_at":  now,
		})

		// 只有 to-web 消息才推送到前端
		if shouldPushToWeb {
			h.hub.SendToUser(userID, "reply", map[string]interface{}{
				"reply":       displayText,
				"reply_audio": audioURL,
			})
		}
	}()
}

// GetReply 获取最近发送消息的回复
// 逻辑：找到最新发送的消息，检查它是否已有回复
func (h *UploadHandler) GetReply(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 获取用户最新发送的一条消息（不管状态）
	var latestMessage model.Message
	err := h.db.Where("user_id = ?", userID).
		Order("sent_at desc").
		First(&latestMessage).Error

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "no_message",
		})
		return
	}

	// 检查这条消息的状态
	switch latestMessage.Status {
	case "replied":
		c.JSON(http.StatusOK, gin.H{
			"status":      "ready",
			"message_id":  latestMessage.ID,
			"text":        latestMessage.Text,
			"reply":       latestMessage.Reply,
			"reply_audio": latestMessage.ReplyAudio,
			"replied_at":  latestMessage.RepliedAt,
		})
	case "timeout":
		c.JSON(http.StatusOK, gin.H{
			"status":     "timeout",
			"message_id": latestMessage.ID,
			"text":       latestMessage.Text,
		})
	default: // "sent" 或其他
		c.JSON(http.StatusOK, gin.H{
			"status":     "waiting",
			"message_id": latestMessage.ID,
			"text":       latestMessage.Text,
		})
	}
}

// GetHistory 获取用户的对话历史（最近N条）
func (h *UploadHandler) GetHistory(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 获取最近3条消息（按创建时间倒序，最新的在前）
	var messages []model.Message
	err := h.db.Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(3).
		Find(&messages).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取历史记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}
