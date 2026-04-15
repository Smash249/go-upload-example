package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
	"upload/db"
	"upload/db/model"
	"upload/oss"
	"upload/read"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GenerateImageRequest struct {
	apiKey string
	apiURL string
	Reader *read.ResourceReader
	Oss    *oss.OSSClient
	Db     *db.CustomDB
	utils  *RequestUtils
	config *RequestConfig

	Concurrency int
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ResponseBody struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message ChoiceMessage `json:"message"`
}

type ChoiceMessage struct {
	Content string `json:"content"`
}

type task struct {
	Data  read.XlsxData
	Ratio string
}

type SaveMode int

const (
	SaveToLocal SaveMode = iota
	SaveToOSS
	SaveToOSSWebp
)

func NewGenerateImageRequest() (*GenerateImageRequest, error) {
	dbClient, err := db.NewCustomDB()
	defer func() {
		if dbClient != nil {
			if err := dbClient.Close(); err != nil {
				fmt.Printf("数据库连接关闭失败: %v\n", err)
			}
		}
	}()
	if err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		dbClient = nil
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	return &GenerateImageRequest{
		apiKey:      os.Getenv("SORA_APIKey"),
		apiURL:      os.Getenv("SORA_APIURL"),
		Db:          dbClient,
		Reader:      read.NewResourceReader(),
		Oss:         oss.NewOSSClient(),
		utils:       NewRequestUtils(),
		config:      NewRequestConfig(),
		Concurrency: 5,
	}, nil
}

// callAPI 调用 AI 接口，返回图片 URL 列表
func (g *GenerateImageRequest) callAPI(data read.XlsxData, ratio string) ([]string, error) {
	prompt := g.utils.buildPrompt(data)

	validRatios := map[string]bool{
		"2:3": true, "3:2": true, "1:1": true,
	}
	if ratio != "" && validRatios[ratio] {
		prompt = fmt.Sprintf("%s【%s】", prompt, ratio)
	}

	payload := RequestBody{
		Model: "sora_image",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("请求体序列化失败: %w", err)
	}

	req, err := http.NewRequest("POST", g.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d，响应: %s", resp.StatusCode, string(bodyBytes))
	}

	var result ResponseBody
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w，原始响应: %s", err, string(bodyBytes))
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("响应没有 choices")
	}

	imageURLs := g.utils.extractImageURLs(result.Choices[0].Message.Content)
	if len(imageURLs) == 0 {
		return nil, fmt.Errorf("未从响应中提取到图片链接")
	}

	return imageURLs, nil
}

// SaveLocal 下载图片并保存到本地
func (g *GenerateImageRequest) SaveLocal(data read.XlsxData, imageURL string) (string, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()

	savePath, err := g.utils.getSavePath()
	if err != nil {
		return "", fmt.Errorf("获取保存路径失败: %w", err)
	}

	ext := g.utils.getExtFromURL(imageURL)
	timestamp := time.Now().UTC().UnixMilli()
	safeName := g.utils.sanitizeFileName(data.BaseText)
	fileName := fmt.Sprintf("%s_%d%s", safeName, timestamp, ext)
	filePath := filepath.Join(savePath, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("保存图片失败: %w", err)
	}

	return filePath, nil
}

// DownloadAndUploadAsWebp 下载图片，内存中转 webp
func (g *GenerateImageRequest) DownloadAndUploadAsWebp(data read.XlsxData, imageURL string) (string, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()

	timestamp := time.Now().UTC().UnixMilli()
	safeName := g.utils.sanitizeFileName(data.BaseText)
	ext := ".webp"
	objectKey := fmt.Sprintf("wordImages/%s_%d%s",
		safeName,
		timestamp,
		ext,
	)
	err = g.Oss.UploadAsWebp(objectKey, resp.Body, 60)
	if err != nil {
		return "", fmt.Errorf("压缩上传失败: %w", err)
	}

	return objectKey, nil
}

// DownloadAndUploadToOSS 下载图片直接流式上传到 OSS
func (g *GenerateImageRequest) DownloadAndUploadToOSS(data read.XlsxData, imageURL string) (string, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()
	ext := g.utils.getExtFromURL(imageURL)
	timestamp := time.Now().UTC().UnixMilli()
	safeName := g.utils.sanitizeFileName(data.BaseText)
	objectKey := fmt.Sprintf("wordImages/%d_%s%s",
		timestamp,
		safeName,
		ext,
	)

	err = g.Oss.UploadFromReader(objectKey, resp.Body)
	if err != nil {
		return "", fmt.Errorf("上传到 OSS 失败: %w", err)
	}
	return objectKey, nil
}

// GenerateImage 生成图片
func (g *GenerateImageRequest) GenerateImage(data read.XlsxData, ratio string, mode SaveMode) (string, error) {
	imageURLs, err := g.callAPI(data, ratio)
	if err != nil {
		return "", err
	}
	switch mode {
	case SaveToOSS:
		return g.DownloadAndUploadToOSS(data, imageURLs[0])
	case SaveToOSSWebp:
		return g.DownloadAndUploadAsWebp(data, imageURLs[0])
	default:
		return g.SaveLocal(data, imageURLs[0])
	}
}

func (g *GenerateImageRequest) saveToDB(data read.XlsxData, imageURL string) error {
	if g.Db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	categoryID, _ := primitive.ObjectIDFromHex(g.config.GetCategoryId())
	seriesKey, ok := g.config.GetSeriesId(data.SeriesName)
	if !ok {
		return fmt.Errorf("未找到系列ID，系列名称: %s", data.SeriesName)
	}
	seriesID, _ := primitive.ObjectIDFromHex(seriesKey)
	return g.Db.InsertOrUpdate("words", model.Word{
		CategoryID:          categoryID,
		SeriesID:            seriesID,
		BaseText:            data.BaseText,
		PosText:             data.PosText,
		PronText:            data.PronText,
		DefinitionText:      data.DefinitionText,
		LearnerExamplesText: data.LearnerExamples,
		ChineseMeaning:      data.ChineseMeaning,
		Image:               imageURL,
		Sort:                0,
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	})
}

// handelExcute 维护一个固定大小的 Worker Pool 并发执行任务
func (g *GenerateImageRequest) handelExcute(data []read.XlsxData, mode SaveMode) {
	if len(data) == 0 {
		return
	}
	maxWorkers := g.Concurrency
	if len(data) < maxWorkers {
		maxWorkers = len(data)
	}

	taskCh := make(chan task, maxWorkers)

	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for t := range taskCh {
				time.Sleep(1 * time.Second)
				fmt.Printf("[Worker %d] 开始处理: %s\n", workerID, t.Data.BaseText)
				imageURL, err := g.GenerateImage(t.Data, t.Ratio, mode)
				if err != nil {
					fmt.Printf("[Worker %d] 生成图片失败 (%s): %v\n", workerID, t.Data.BaseText, err)
					continue
				}

				if err := g.saveToDB(t.Data, imageURL); err != nil {
					fmt.Printf("[Worker %d] 保存到数据库失败 (%s): %v\n", workerID, t.Data.BaseText, err)
				} else {
					fmt.Printf("[Worker %d] 处理完成: %s -> %s\n", workerID, t.Data.BaseText, imageURL)
				}
			}
		}(i)
	}
	for _, item := range data {
		taskCh <- task{Data: item}
	}
	close(taskCh)
	wg.Wait()
}

// run 基础启动器
func (g *GenerateImageRequest) run(mode SaveMode) error {
	dataMap, err := g.Reader.Run()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for seriesName, dataList := range dataMap {
		wg.Add(1)
		fmt.Printf("开始处理系列 -> %s (共 %d 条)\n", seriesName, len(dataList))
		go func(name string, list []read.XlsxData) {
			defer wg.Done()
			g.handelExcute(list, mode)
			fmt.Printf("系列处理完成 -> %s\n", name)
		}(seriesName, dataList)
	}
	wg.Wait()

	fmt.Println("全部系列处理完成")
	return nil
}

// Run 生成图片并保存到本地
func (g *GenerateImageRequest) Run() error {
	return g.run(SaveToLocal)
}

// RunWithOSS 生成图片并上传到 OSS（原格式）
func (g *GenerateImageRequest) RunWithOSS() error {
	return g.run(SaveToOSS)
}

// RunWithOSSWebp 生成图片并转 webp 上传到 OSS
func (g *GenerateImageRequest) RunWithOSSWebp() error {
	return g.run(SaveToOSSWebp)
}
