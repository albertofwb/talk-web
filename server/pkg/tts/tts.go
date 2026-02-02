package tts

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type TTS struct {
	ScriptPath string
	Timeout    time.Duration
}

func NewTTS() *TTS {
	return &TTS{
		ScriptPath: "/home/albert/.local/bin/xiaoxiao-tts",
		Timeout:    30 * time.Second,
	}
}

// Generate 生成语音文件（自动路径）
func (t *TTS) Generate(text string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, t.ScriptPath, text)
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("TTS timeout after %v", t.Timeout)
		}
		return "", fmt.Errorf("TTS generation failed: %w", err)
	}

	// 脚本输出文件路径
	filePath := strings.TrimSpace(string(output))
	if filePath == "" {
		return "", fmt.Errorf("TTS script returned empty path")
	}

	return filePath, nil
}

// GenerateWithPath 生成到指定路径
func (t *TTS) GenerateWithPath(text, outputPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, t.ScriptPath, text, outputPath)
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("TTS timeout after %v", t.Timeout)
		}
		return fmt.Errorf("TTS generation failed: %w", err)
	}

	return nil
}
