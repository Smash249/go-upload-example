package request

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"upload/read"
)

type RequestUtils struct{}

func NewRequestUtils() *RequestUtils {
	return &RequestUtils{}
}

func (ru *RequestUtils) getSavePath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	savePath := filepath.Join(wd, "download")
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		if err := os.MkdirAll(savePath, os.ModePerm); err != nil {
			return "", fmt.Errorf("创建保存目录失败: %w", err)
		}
	}
	return savePath, nil
}

func (ru *RequestUtils) getExtFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ".png"
	}
	ext := filepath.Ext(u.Path)
	if ext == "" {
		return ".png"
	}
	return ext
}

func (ru *RequestUtils) sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_",
		"|", "_", ",", "_", " ", "_",
	)
	name = replacer.Replace(name)
	re := regexp.MustCompile(`_+`)
	name = re.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	return name
}

func (ru *RequestUtils) extractImageURLs(content string) []string {
	re := regexp.MustCompile(`!\[.*?\]\((https?://[^)]+)\)`)
	matches := re.FindAllStringSubmatch(content, -1)
	urls := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			urls = append(urls, match[1])
		}
	}
	return urls
}

func (ru *RequestUtils) getSystemPrompt() string {
	return "你是一个图像生成助手。请根据用户提供的词语、定义和学习者例句等信息，创作出符合要求的图片。请确保图片内容与提供的信息相关，并且具有创意和吸引力。图片中不要包含单词本身，而是要通过视觉元素来表达词语的含义和相关信息。让学习者能够通过图片更好地理解和记忆词语的意义和用法。"
}

func (ru *RequestUtils) buildPrompt(data read.XlsxData) string {
	basePrompt := ru.getSystemPrompt()
	if data.BaseText != "" {
		basePrompt += fmt.Sprintf(" 词语: %s;", data.BaseText)
	}
	if data.DefinitionText != "" {
		basePrompt += fmt.Sprintf(" 定义: %s;", data.DefinitionText)
	}
	if data.LearnerExamples != "" {
		example := strings.Split(data.LearnerExamples, ";")[0]
		basePrompt += fmt.Sprintf(" 学习者例句: %s;", example)
	}
	return basePrompt
}
