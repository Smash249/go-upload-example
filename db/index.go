package db

import (
	"context"
	"fmt"
	"os"
	"time"
	"upload/db/model"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CustomDBConfig struct {
	URI    string
	DBName string
}

type CustomDB struct {
	Client *mongo.Client
	DB     *mongo.Database
	CustomDBConfig
}

func NewCustomDB() (*CustomDB, error) {
	config := CustomDBConfig{
		URI:    os.Getenv("MONGO_URI"),
		DBName: os.Getenv("MONGO_DB"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.URI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	db := client.Database(config.DBName)
	fmt.Println("数据库连接成功")
	return &CustomDB{
		DB:             db,
		CustomDBConfig: config,
	}, nil
}

func (c *CustomDB) GetCollection(collectionName string) *mongo.Collection {
	return c.DB.Collection(collectionName)
}

func (c *CustomDB) InsertOne(collectionName string, document model.Word) error {
	collection := c.GetCollection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, document)
	return err
}

func (c *CustomDB) Close() error {
	if c.Client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return c.Client.Disconnect(ctx)
}
