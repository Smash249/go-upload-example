package audio

import (
	"bytes"
	"context"
	"fmt"

	api "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/speak/v1/rest"
	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	client "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/speak"
)

type DeepgramClient struct{}

func NewDeepgramClient() *DeepgramClient {
	client.InitWithDefault()
	return &DeepgramClient{}
}

func (d *DeepgramClient) TextToSpeech(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("文本内容不能为空")
	}

	options := &interfaces.SpeakOptions{
		Model:    "aura-2-thalia-en",
		Encoding: "mp3",
	}

	c := client.NewRESTWithDefaults()
	dg := api.New(c)

	var buffer interfaces.RawResponse

	_, err := dg.ToStream(ctx, text, options, &buffer)
	if err != nil {
		return nil, fmt.Errorf("Deepgram TTS 请求失败: %w", err)
	}

	data := buffer.Bytes()
	if len(data) == 0 {
		return nil, fmt.Errorf("Deepgram 返回音频数据为空")
	}

	return data, nil
}

func (d *DeepgramClient) TextToSpeechReader(ctx context.Context, text string) (*bytes.Reader, error) {
	data, err := d.TextToSpeech(ctx, text)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}
