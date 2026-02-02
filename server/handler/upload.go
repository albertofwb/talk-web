package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"talk-web/server/pkg/stt"
	"talk-web/server/pkg/tts"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	stt *stt.STT
	tts *tts.TTS
}

func NewUploadHandler(talkServerURL string) *UploadHandler {
	return &UploadHandler{
		stt: stt.NewSTT(),
		tts: tts.NewTTS(),
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

	// 生成回复文本
	replyText := fmt.Sprintf("收到 %s 的消息：%s", username, recognizedText)

	// TTS: 文字转语音
	replyAudioPath, err := h.tts.Generate(replyText)
	if err != nil {
		// TTS 失败不影响返回识别结果
		c.JSON(http.StatusOK, gin.H{
			"text":     recognizedText,
			"reply":    replyText,
			"user_id":  userID,
			"username": username,
			"tts_error": err.Error(),
		})
		return
	}

	// 返回识别结果和回复语音 URL
	audioFilename := filepath.Base(replyAudioPath)
	audioURL := fmt.Sprintf("/api/audio/%s", audioFilename)

	c.JSON(http.StatusOK, gin.H{
		"text":       recognizedText,
		"reply":      replyText,
		"reply_audio": audioURL,
		"user_id":    userID,
		"username":   username,
	})
}
