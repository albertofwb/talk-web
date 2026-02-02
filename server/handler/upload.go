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

		// TTS: 文字转语音
		replyAudioPath, err := h.tts.Generate(replyText)
		audioURL := ""
		if err != nil {
			fmt.Printf("[TTS Error] Failed to generate: %v\n", err)
		} else {
			audioFilename := filepath.Base(replyAudioPath)
			audioURL = fmt.Sprintf("/api/audio/%s", audioFilename)
			fmt.Printf("[TTS Success] Audio generated: %s\n", audioFilename)
		}

		// 更新数据库记录
		now := time.Now()
		h.db.Model(&message).Updates(map[string]interface{}{
			"reply":       replyText,
			"reply_audio": audioURL,
			"status":      "replied",
			"replied_at":  now,
		})

		// 通过 WebSocket 推送回复给用户
		h.hub.SendToUser(userID, "reply", map[string]interface{}{
			"reply":       replyText,
			"reply_audio": audioURL,
		})
	}()
}

// GetReply 轮询获取回复（非阻塞）
func (h *UploadHandler) GetReply(c *gin.Context) {
	username := c.GetString("username")

	// 检查是否有新消息（不弹出）
	msg, err := h.tg.GetFromTelegram(telegram.DefaultUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取消息失败",
			"detail": err.Error(),
		})
		return
	}

	// 没有消息
	if msg == nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "waiting",
			"username": username,
		})
		return
	}

	// 有消息，弹出它
	msg, err = h.tg.PopFromTelegram(telegram.DefaultUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "读取消息失败",
		})
		return
	}

	replyText := msg.Text

	// 查找最近一条已回复的消息，获取音频 URL
	var latestMessage model.Message
	err = h.db.Where("user_id = ? AND status = 'replied' AND reply = ?",
		c.GetUint("user_id"), replyText).
		Order("replied_at desc").
		First(&latestMessage).Error

	replyAudio := ""
	if err == nil && latestMessage.ReplyAudio != "" {
		replyAudio = latestMessage.ReplyAudio
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "ready",
		"reply":       replyText,
		"reply_audio": replyAudio,
		"sender":      msg.Sender,
		"timestamp":   msg.Timestamp,
	})
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
