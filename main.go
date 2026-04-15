package main

import (
	"context"
	"fmt"
	"log"
	"upload/audio"
	"upload/db"
	"upload/request"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("加载 .env 失败")
	}
	generateAudio()
}

func generateImage() {
	generator, err := request.NewGenerateImageRequest()
	if err != nil {
		log.Fatalf("创建 GenerateImageRequest 失败: %v", err)
	}
	if err := generator.RunWithOSSWebp(); err != nil {
		log.Fatalf("运行失败: %v", err)
	}
}

func generateAudio() {
	database, err := db.NewCustomDB()
	if err != nil {
		panic(fmt.Sprintf("连接数据库失败: %v", err))
	}
	defer database.Close()
	//   69c6380e2cdb56c2fcc5f0ed => a1
	//   69c6533bb47417171cb4642a => a2
	//   69d493000abeb5c5ebc74cb7 => ket
	//   69ddb50e6972ef48c146ecbb => awl
	seriesId := "69ddb50e6972ef48c146ecbb"
	err = audio.NewCustomAudio(database).GenerateAudioBySeriesID(context.Background(), seriesId)
	if err != nil {
		fmt.Printf("系列ID %s 处理失败: %v\n", seriesId, err)
		return
	}
	fmt.Printf("系列ID %s 处理完成\n", seriesId)
}
