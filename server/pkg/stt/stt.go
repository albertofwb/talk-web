package stt

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type STT struct {
	ScriptPath string
	ModelSize  string
	Timeout    time.Duration
}

func NewSTT() *STT {
	return &STT{
		ScriptPath: "/home/albert/.local/bin/stt",
		ModelSize:  "base",
		Timeout:    60 * time.Second,
	}
}

// Transcribe 语音转文字
func (s *STT) Transcribe(audioPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.ScriptPath, audioPath, "-m", s.ModelSize)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("STT timeout after %v", s.Timeout)
		}
		return "", fmt.Errorf("STT failed: %w (output: %s)", err, string(output))
	}

	text := strings.TrimSpace(string(output))
	if text == "" {
		return "", fmt.Errorf("no text recognized")
	}

	return text, nil
}

// TranscribeWithModel 使用指定模型识别
func (s *STT) TranscribeWithModel(audioPath, modelSize string) (string, error) {
	oldModel := s.ModelSize
	s.ModelSize = modelSize
	defer func() { s.ModelSize = oldModel }()
	return s.Transcribe(audioPath)
}
