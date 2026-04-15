package db

import (
	"context"
	"fmt"
	"time"
	"upload/db/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const wordCollection = "words"

// FindWordsBySeriesID 根据系列 ID 查询该系列下的所有单词
func (c *CustomDB) FindWordsBySeriesID(seriesID primitive.ObjectID) ([]model.Word, error) {
	collection := c.GetCollection(wordCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*60*time.Second)
	defer cancel()

	filter := bson.M{"seriesId": seriesID}
	page := int64(7)       // 第几页，从 1 开始  ket 1800 awl 3101
	pageSize := int64(500) // 每页 500 条

	skip := (page - 1) * pageSize

	opts := options.Find().
		SetLimit(pageSize).
		SetSkip(skip)
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询单词列表失败: %w", err)
	}
	defer cursor.Close(ctx)

	var words []model.Word
	if err := cursor.All(ctx, &words); err != nil {
		return nil, fmt.Errorf("解析单词列表失败: %w", err)
	}
	return words, nil
}

// UpdateWordAudio 更新单词的 audio 字段
func (c *CustomDB) UpdateWordAudio(wordID primitive.ObjectID, audioPath string) error {
	collection := c.GetCollection(wordCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": wordID}
	update := bson.M{
		"$set": bson.M{
			"audio":     audioPath,
			"updatedAt": time.Now(),
		},
	}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新单词音频路径失败: %w", err)
	}
	return nil
}
