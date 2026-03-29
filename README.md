# go-upload-example

一个基于 Go 语言开发的词汇图片自动生成与上传工具。该工具从 Excel 文件中批量读取单词数据，通过 Sora AI 接口自动生成对应的配图，将图片转换为 WebP 格式后上传至阿里云 OSS，并将词汇及图片信息写入 MongoDB 数据库。

---

## 功能特性

- **批量读取 Excel**：自动扫描 `resources/` 目录下的所有 `.xlsx` 文件，解析词汇信息（单词、词性、音标、释义、例句等）
- **AI 图片生成**：调用 Sora AI 图像生成接口，根据词汇内容自动构建提示词并生成配图
- **WebP 格式转换**：在内存中将原始图片无损转换为 WebP 格式（可自定义压缩质量），减小存储体积
- **阿里云 OSS 上传**：将处理后的图片流式上传到阿里云 OSS，存储于 `wordImages/` 路径下
- **MongoDB 数据持久化**：将词汇元信息与 OSS 图片地址一并写入 MongoDB 数据库
- **并发处理**：内置 Worker Pool 并发机制（默认 5 个 Worker），支持多系列并行处理，大幅提升批量处理效率

---

## 项目结构

```
go-upload-example/
├── main.go                 # 程序入口，加载 .env 并启动任务
├── go.mod / go.sum         # Go 模块依赖管理
├── makefile                # 快捷运行命令
├── envExample              # 环境变量配置示例
├── resources/              # 待处理的 Excel 数据文件
│   ├── a1.xlsx             # 系列 a1 的词汇数据
│   ├── a2.xlsx             # 系列 a2 的词汇数据
│   ├── b1.xlsx             # 系列 b1 的词汇数据
│   ├── b2.xlsx             # 系列 b2 的词汇数据
│   ├── c1.xlsx             # 系列 c1 的词汇数据
│   └── c2.xlsx             # 系列 c2 的词汇数据
├── request/                # 核心业务逻辑
│   ├── index.go            # 图片生成、下载、上传、入库的主流程
│   ├── config.go           # 系列 ID 与分类 ID 配置
│   └── utils.go            # 提示词构建、文件名处理、URL 解析等工具函数
├── read/                   # Excel 文件读取模块
│   └── index.go            # 解析 .xlsx 并返回词汇结构化数据
├── oss/                    # 阿里云 OSS 操作模块
│   └── index.go            # 上传文件、WebP 转换、生成签名 URL
└── db/                     # MongoDB 数据库模块
    ├── index.go            # 数据库连接、集合操作
    └── model/
        └── word.go         # Word 数据模型定义
```

---

## 数据流程

```
resources/*.xlsx
      │
      ▼
  read 模块（解析词汇数据）
      │
      ▼
  request 模块（构建 AI 提示词）
      │
      ▼
  Sora AI API（生成图片 URL）
      │
      ▼
  oss 模块（下载 → WebP 转换 → 上传到阿里云 OSS）
      │
      ▼
  db 模块（写入 MongoDB words 集合）
```

---

## Excel 数据格式

`resources/` 目录下的 `.xlsx` 文件**第一行为表头**，从第二行起为数据，列顺序如下：

| 列索引 | 字段           | 说明                   |
|--------|----------------|------------------------|
| 0      | `baseText`     | 单词                   |
| 1      | `posText`      | 词性                   |
| 2      | —              | 保留列，程序不读取     |
| 3      | `pronText`     | 音标                   |
| 4      | `definitionText` | 释义                 |
| 5      | `learnerExamplesText` | 学习者例句       |

> 文件名（去掉 `.xlsx` 扩展名）即为系列名称（如 `a1`、`b2`），需在 `request/config.go` 中配置对应的系列 ID。

---

## 环境准备

### 依赖要求

- Go 1.21+
- 阿里云 OSS Bucket（已开通公网访问）
- MongoDB 实例（可使用 Atlas 或自建）
- Sora AI API 账号（需 API Key 及接口地址）

### 安装依赖

```bash
go mod download
```

### 配置环境变量

复制 `envExample` 为 `.env` 并填写各项配置：

```bash
cp envExample .env
```

编辑 `.env` 文件：

```dotenv
# Sora AI 图像生成接口
SORA_APIKey=你的SORA_APIKey
SORA_APIURL=你的SORA_APIURL

# MongoDB 连接信息
MONGO_URI=你的MONGO_URI
MONGO_DB=你的MONGO_DB

# 阿里云 OSS 配置
AccessKeyID=你的阿里云AccessKeyID
AccessKeySecret=你的阿里云AccessKeySecret
Endpoint=你的阿里云Endpoint
BucketName=你的阿里云BucketName
```

---

## 系列配置

在 `request/config.go` 中维护了分类 ID 与系列 ID 的映射关系，需与 MongoDB 中的实际文档 ID 保持一致：

```go
func NewRequestConfig() *RequestConfig {
    return &RequestConfig{
        categoryId: "69c637e22cdb56c2fcc5f0df",  // 分类 ID
        mp: map[string]string{
            "a1": "69c6380e2cdb56c2fcc5f0ed",
            "a2": "69c6533bb47417171cb4642a",
            "b1": "69c65362b47417171cb4643d",
            "b2": "69c65376b47417171cb46443",
            "c1": "69c65396b47417171cb46449",
            "c2": "69c653adb47417171cb4644f",
        },
    }
}
```

如需新增系列，在 `resources/` 目录中放入对应 `.xlsx` 文件，并在此处添加相应的系列 ID 映射即可。

---

## 运行

```bash
# 使用 Makefile（推荐）
make run

# 或直接运行
go run main.go
```

程序启动后将自动：
1. 扫描 `resources/` 目录下的所有 `.xlsx` 文件
2. 按系列分组，并发处理每个系列的词汇条目
3. 对每个词汇调用 AI 接口生成图片，转换为 WebP 后上传至 OSS
4. 将词汇数据及图片地址写入 MongoDB `words` 集合

### 运行模式

`request/index.go` 中提供了三种运行模式，可根据需要在 `main.go` 中切换：

| 方法                | 说明                              |
|---------------------|-----------------------------------|
| `Run()`             | 生成图片并保存到本地 `download/` 目录 |
| `RunWithOSS()`      | 生成图片并以原格式上传到 OSS       |
| `RunWithOSSWebp()`  | 生成图片，转换为 WebP 后上传到 OSS（默认） |

---

## 主要依赖

| 依赖包 | 用途 |
|--------|------|
| [aliyun/aliyun-oss-go-sdk](https://github.com/aliyun/aliyun-oss-go-sdk) | 阿里云 OSS 文件上传 |
| [chai2010/webp](https://github.com/chai2010/webp) | 图片 WebP 格式编解码 |
| [joho/godotenv](https://github.com/joho/godotenv) | 读取 `.env` 配置文件 |
| [xuri/excelize](https://github.com/xuri/excelize) | 读取 Excel `.xlsx` 文件 |
| [mongo-driver](https://github.com/mongodb/mongo-go-driver) | MongoDB 官方 Go 驱动 |
| [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) | 扩展图片格式支持（BMP、TIFF、WebP） |

---

## 注意事项

- `.env` 文件包含敏感凭据，**请勿提交至版本控制**（已在 `.gitignore` 中排除）
- Excel 文件的列顺序需严格按照上述格式，否则会导致数据解析错误
- 系列名称（Excel 文件名）需在 `request/config.go` 中存在对应映射，否则写库时会报错
- 并发数默认为 5，可通过修改 `GenerateImageRequest.Concurrency` 字段调整
