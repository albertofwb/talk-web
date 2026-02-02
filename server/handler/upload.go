package handler

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	talkServerURL string
}

func NewUploadHandler(talkServerURL string) *UploadHandler {
	return &UploadHandler{
		talkServerURL: talkServerURL,
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

	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	// 转发到 talk-server
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加音频文件
	part, err := writer.CreateFormFile("audio", header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建表单失败"})
		return
	}
	if _, err := part.Write(fileBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "写入文件失败"})
		return
	}

	// 添加用户信息
	writer.WriteField("user_id", fmt.Sprintf("%d", userID))
	writer.WriteField("username", username)
	writer.Close()

	// 发送到 talk-server
	req, err := http.NewRequest("POST", h.talkServerURL+"/stt", &buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "转发到talk-server失败"})
		return
	}
	defer resp.Body.Close()

	// 读取 talk-server 的响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取响应失败"})
		return
	}

	// 返回结果
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}
