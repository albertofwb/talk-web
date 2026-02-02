package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	// 获取消息ID（前端生成的唯一标识）
	msgID := c.PostForm("msg_id")
	if msgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 msg_id 参数"})
		return
	}

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
		MessageID: msgID,  // 添加消息ID
		UserID:    userID,
		Username:  username,
		Text:      recognizedText,
		Status:    "sent",
		SentAt:    time.Now(),
	}
	if err := h.db.Create(&message).Error; err != nil {
		fmt.Printf("[DB Error] Failed to save message: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存消息失败"})
		return
	}

	// 新格式：from-web:[user_id]:[msg_id] 消息内容
	telegramText := fmt.Sprintf("from-web:%d:%s %s", userID, msgID, recognizedText)
	fmt.Printf("[Format] Telegram message: %s\n", telegramText)

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
		"text":       recognizedText,
		"message_id": msgID,  // 返回消息ID
		"status":     "sent",
		"message":    "消息已发送，等待回复中...",
		"user_id":    userID,
		"username":   username,
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

		// 解析新格式：to-web:[user_id]:[msg_id] 回复内容
		// 检查是否为 to-web 消息
		if len(replyText) < 7 || replyText[:7] != "to-web:" {
			fmt.Printf("[Telegram] 非 to-web 消息，忽略\n")
			h.db.Model(&message).Update("status", "ignored")
			return
		}

		// 分割 header 和 content
		parts := strings.SplitN(replyText, " ", 2)
		if len(parts) < 2 {
			fmt.Printf("[Telegram Error] 格式错误，缺少消息内容\n")
			return
		}

		header := parts[0]      // "to-web:[user_id]:[msg_id]"
		displayText := parts[1] // "回复内容"

		// 解析 header
		headerParts := strings.Split(header, ":")
		if len(headerParts) != 3 {
			fmt.Printf("[Telegram Error] Header 格式错误: %s\n", header)
			return
		}

		replyUserIDStr := headerParts[1]
		replyMsgID := headerParts[2]

		// 验证 user_id 匹配
		var replyUserID uint
		fmt.Sscanf(replyUserIDStr, "%d", &replyUserID)
		if replyUserID != userID {
			fmt.Printf("[Telegram Warning] UserID 不匹配: 期望 %d, 实际 %d\n", userID, replyUserID)
			// 不推送给当前用户
			return
		}

		// 验证 msg_id 匹配
		if replyMsgID != msgID {
			fmt.Printf("[Telegram Warning] MessageID 不匹配: 期望 %s, 实际 %s\n", msgID, replyMsgID)
			// 不推送给当前用户
			return
		}

		fmt.Printf("[Telegram] ✓ 验证通过 - UserID: %d, MessageID: %s\n", userID, msgID)

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

		// 推送到前端（已验证 user_id 和 msg_id 匹配）
		h.hub.SendToUser(userID, "reply", map[string]interface{}{
			"message_id":  msgID,
			"reply":       displayText,
			"reply_audio": audioURL,
		})
		fmt.Printf("[WebSocket] 推送回复给用户 %d, 消息ID: %s\n", userID, msgID)
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
