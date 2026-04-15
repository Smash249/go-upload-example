package audio

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
	"upload/db"
	"upload/oss"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type customAudio struct {
	db         *db.CustomDB
	oss        *oss.OSSClient
	deepgram   *DeepgramClient
	elevenlabs *ElevenlabsClient
	custom     *CustomClient
}

func NewCustomAudio(database *db.CustomDB) *customAudio {
	return &customAudio{
		db:         database,
		oss:        oss.NewOSSClient(),
		deepgram:   NewDeepgramClient(),
		elevenlabs: NewElevenlabsClient(),
		custom:     NewCustomClient(),
	}
}

// GenerateAudioBySeriesID 根据系列ID批量生成单词音频并上传到OSS，保存路径到数据库
func (c *customAudio) GenerateAudioBySeriesID(ctx context.Context, seriesID string) error {
	objID, err := primitive.ObjectIDFromHex(seriesID)
	if err != nil {
		return fmt.Errorf("无效的系列ID: %w", err)
	}
	words, err := c.db.FindWordsBySeriesID(objID)
	if err != nil {
		return fmt.Errorf("查询单词失败: %w", err)
	}
	if len(words) == 0 {
		return fmt.Errorf("该系列下没有找到任何单词")
	}
	successCount := 0
	failCount := 0
	for i, word := range words {
		fmt.Printf("[%d/%d] 正在处理: %s\n", i+1, len(words), word.BaseText)
		if word.Audio != "" {
			fmt.Printf("  删除旧音频: %s\n", word.Audio)
			err = c.oss.DeleteObject(word.Audio)
			if err != nil {
				fmt.Printf("  删除旧音频失败（继续执行）: %v\n", err)
			}
		}
		audioData, err := c.custom.TextToSpeech(ctx, word.BaseText)
		if err != nil {
			fmt.Printf("  生成音频失败: %v\n", err)
			failCount++
			continue
		}

		fileName := strings.NewReplacer(" ", "_", "/", "_").Replace(word.BaseText)
		objectKey := fmt.Sprintf("wordAudios/%s_%d.mp3",
			fileName,
			time.Now().UnixMilli(),
		)
		err = c.oss.UploadFromReader(objectKey, bytes.NewReader(audioData))
		if err != nil {
			fmt.Printf("  上传OSS失败: %v\n", err)
			failCount++
			continue
		}
		err = c.db.UpdateWordAudio(word.ID, objectKey)
		if err != nil {
			fmt.Printf("  更新数据库失败: %v\n", err)
			failCount++
			continue
		}
		fmt.Printf("  成功: %s\n", objectKey)
		successCount++
	}

	fmt.Printf("\n处理完成！成功: %d, 失败: %d, 总计: %d\n", successCount, failCount, len(words))
	return nil
}
