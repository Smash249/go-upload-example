package audio

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	elevenlabs "github.com/haguro/elevenlabs-go"
)

type ElevenlabsClient struct {
	client  *elevenlabs.Client
	voiceID string
	modelID string
}

func NewElevenlabsClient() *ElevenlabsClient {
	client := elevenlabs.NewClient(context.Background(), os.Getenv("ELEVENLABS_API_KEY"), 30*time.Second)
	return &ElevenlabsClient{
		client:  client,
		voiceID: os.Getenv("ELEVENLABS_VOICE_ID"),
		modelID: os.Getenv("ELEVENLABS_MODEL_ID"),
	}
}

func (e *ElevenlabsClient) TextToSpeech(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("文本内容不能为空")
	}

	ttsReq := elevenlabs.TextToSpeechRequest{
		Text:    text,
		ModelID: e.modelID,
	}

	data, err := e.client.TextToSpeech(e.voiceID, ttsReq)
	if err != nil {
		return nil, fmt.Errorf("ElevenLabs TTS 请求失败: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("ElevenLabs 返回音频数据为空")
	}

	return data, nil
}

func (e *ElevenlabsClient) TextToSpeechReader(ctx context.Context, text string) (*bytes.Reader, error) {
	data, err := e.TextToSpeech(ctx, text)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}
