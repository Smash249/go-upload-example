package main

import (
	"log"
	"upload/request"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("加载 .env 失败")
	}
	genertor, err := request.NewGenerateImageRequest()
	if err != nil {
		log.Fatalf("创建 GenerateImageRequest 失败: %v", err)
	}
	if err := genertor.RunWithOSSWebp(); err != nil {
		log.Fatalf("运行失败: %v", err)
	}
}
