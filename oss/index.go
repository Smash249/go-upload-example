package oss

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/chai2010/webp"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
}

type OSSClient struct {
	Config OSSConfig
}

func NewOSSClient() *OSSClient {
	return &OSSClient{Config: OSSConfig{
		Endpoint:        os.Getenv("Endpoint"),
		AccessKeyID:     os.Getenv("AccessKeyID"),
		AccessKeySecret: os.Getenv("AccessKeySecret"),
		BucketName:      os.Getenv("BucketName"),
	}}
}

func (o *OSSClient) getClient() (*oss.Client, error) {
	client, err := oss.New(o.Config.Endpoint, o.Config.AccessKeyID, o.Config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建 OSS 客户端失败: %w", err)
	}
	return client, nil
}

func (o *OSSClient) getBucket() (*oss.Bucket, error) {
	client, err := o.getClient()
	if err != nil {
		return nil, fmt.Errorf("获取 OSS 客户端失败: %w", err)
	}
	bucket, err := client.Bucket(o.Config.BucketName)
	if err != nil {
		return nil, fmt.Errorf("获取 Bucket 失败: %w", err)
	}
	return bucket, nil
}

// UploadAsWebp 从 io.Reader 读取图片，内存中转为 webp 后上传到 OSS
func (o *OSSClient) UploadAsWebp(objectKey string, reader io.Reader, quality float32) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("读取图片数据失败: %w", err)
	}
	contentType := http.DetectContentType(data)
	if len(data) < 100 {
		return fmt.Errorf("图片数据异常，大小: %d bytes, 内容: %s", len(data), string(data))
	}
	var webpData bytes.Buffer
	if contentType == "image/webp" {
		webpData.Write(data)
	} else {
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			img, err = webp.Decode(bytes.NewReader(data))
			if err != nil {
				return fmt.Errorf("解码图片失败, Content-Type: %s, 错误: %w", contentType, err)
			}
		}
		err = webp.Encode(&webpData, img, &webp.Options{Quality: quality})
		if err != nil {
			return fmt.Errorf("编码 webp 失败: %w", err)
		}
	}
	bucket, err := o.getBucket()
	if err != nil {
		return fmt.Errorf("获取 Bucket 失败: %w", err)
	}
	err = bucket.PutObject(objectKey, &webpData)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	return nil
}

// UploadFile 上传本地文件（
func (o *OSSClient) UploadFile(localFilePath string) (string, error) {
	bucket, err := o.getBucket()
	if err != nil {
		return "", fmt.Errorf("获取 Bucket 失败: %w", err)
	}

	file, err := os.Open(localFilePath)
	if err != nil {
		return "", fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer file.Close()

	ext := filepath.Ext(localFilePath)
	objectKey := fmt.Sprintf("wordImages/%s/%d%s",
		time.Now().Format("20060102"),
		time.Now().UnixMilli(),
		ext,
	)
	err = bucket.PutObject(objectKey, file)
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}
	return objectKey, nil
}

// UploadFromReader 从 io.Reader 直接上传
func (o *OSSClient) UploadFromReader(objectKey string, reader io.Reader) error {
	bucket, err := o.getBucket()
	if err != nil {
		return fmt.Errorf("获取 Bucket 失败: %w", err)
	}
	err = bucket.PutObject(objectKey, reader)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}
	return nil
}

// GetSignedURL 生成临时访问链接
func (o *OSSClient) GetSignedURL(objectKey string, expireSeconds int64) (string, error) {
	bucket, err := o.getBucket()
	if err != nil {
		return "", fmt.Errorf("获取 Bucket 失败: %w", err)
	}
	signedURL, err := bucket.SignURL(objectKey, oss.HTTPGet, expireSeconds)
	if err != nil {
		return "", fmt.Errorf("生成签名 URL 失败: %w", err)
	}
	return signedURL, nil
}

func (o *OSSClient) DeleteObject(objectKey string) error {
	bucket, err := o.getBucket()
	if err != nil {
		return fmt.Errorf("获取 Bucket 失败: %w", err)
	}
	err = bucket.DeleteObject(objectKey)
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}
