package audio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type CustomClient struct{}

func NewCustomClient() *CustomClient {
	return &CustomClient{}
}

// https://dict.youdao.com/dictvoice?audio=rading&type=1 => 有道
// https://api.dictionaryapi.dev/api/v2/entries/en/happy =》 Free Dictionary API

func (c *CustomClient) request(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("文本内容不能为空")
	}

	requestUrl := fmt.Sprintf("%s?audio=%s&type=1", "https://dict.youdao.com/dictvoice", url.QueryEscape(text))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求出现错误: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求返回非成功状态码: %d", response.StatusCode)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应数据失败: %w", err)
	}
	return data, nil
}

func (c *CustomClient) TextToSpeech(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("文本内容不能为空")
	}
	data, err := c.request(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("文本转语音失败: %w", err)
	}
	return data, nil
}

func (c *CustomClient) TextToSpeechReader(ctx context.Context, text string) (*bytes.Reader, error) {
	data, err := c.TextToSpeech(ctx, text)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}
