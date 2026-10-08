# OtterIO Go SDK

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.27.1%2B-00ADD8.svg?logo=go)](./go.mod)

[English](./README.md) · 简体中文

OtterIO SDK 是访问 Amazon S3 兼容对象存储的 Go 客户端，可用于连接 [OtterIO](https://github.com/soulteary/otterio)。项目基于 [MinIO Go SDK](https://github.com/minio/minio-go) 分叉。本仓库提供客户端库；[OtterIO 服务端](https://github.com/soulteary/otterio)和 [`oc` 命令行客户端](https://github.com/soulteary/oc)由独立仓库维护。

模块路径是 `github.com/soulteary/otterio-sdk/v7`，Go 包名仍为 `minio`，下方示例通过显式别名导入。迁移现有应用时，需要同时更新根模块和 `pkg/credentials`、`pkg/encrypt` 等辅助包的导入路径。

本项目独立维护，与 MinIO, Inc. 无关联，也未获得其认可或赞助。原始版权声明保留在 [NOTICE](./NOTICE) 和源文件中。

## 安装

按照 [go.mod](./go.mod) 的要求，使用 **Go 1.27.1 或更新版本**。新建一个 Go 模块：

```sh
mkdir otterio-sdk-example
cd otterio-sdk-example
go mod init example.com/otterio-sdk-example
go get github.com/soulteary/otterio-sdk/v7
```

已有模块只需执行 `go get`。需要固定版本时，可在模块路径后加上发布标签，例如 `@v7.3.1`；可用版本见仓库的[标签列表](https://github.com/soulteary/otterio-sdk/tags)。

## 快速开始：上传并读回对象

按 OtterIO 的[快速开始指南](https://github.com/soulteary/otterio/blob/main/README_zh_CN.md#快速开始)启动服务端，或连接已有的 S3 服务。示例所用凭据需要具备检查及创建存储桶、上传和读取对象的权限。程序会在需要时创建 `otterio-sdk-example` 存储桶，并上传 `hello.txt`；再次运行会写入同一个对象键。无需安装存储命令行客户端。

在启动本地 OtterIO 的同一个终端中，复用服务端启动时配置的凭据：

```sh
: "${OTTERIO_ROOT_USER:?请设置启动 OtterIO 时使用的用户名}"
: "${OTTERIO_ROOT_PASSWORD:?请设置启动 OtterIO 时使用的密码}"
export S3_ENDPOINT=127.0.0.1:9000
export S3_ACCESS_KEY="$OTTERIO_ROOT_USER"
export S3_SECRET_KEY="$OTTERIO_ROOT_PASSWORD"
export S3_USE_TLS=false
printf 'Hello from OtterIO SDK!\n' > hello.txt
```

连接其他服务时，设置对应的 `S3_ENDPOINT`、`S3_ACCESS_KEY` 和 `S3_SECRET_KEY`。端点应指向 **S3 API 端口**，不要使用网页控制台端口。`S3_ENDPOINT` 可以采用不含 URL 路径的 `host:port`；示例默认使用 `S3_USE_TLS=true`，即 HTTPS。如果显式包含 `http://` 或 `https://`，协议必须与 `Secure` 一致。上面的回环地址使用 HTTP，因而设置为 `false`；远程连接请使用 HTTPS。根据目标服务调整存储桶名称和 `MakeBucketOptions.Region`。

在模块目录中将以下内容保存为 `main.go`：

```go
package main

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	minio "github.com/soulteary/otterio-sdk/v7"
	"github.com/soulteary/otterio-sdk/v7/pkg/credentials"
)

func requiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("Set %s before running this example", name)
	}
	return value
}

func main() {
	secure := true
	if value := os.Getenv("S3_USE_TLS"); value != "" {
		var err error
		secure, err = strconv.ParseBool(value)
		if err != nil {
			log.Fatal(err)
		}
	}

	client, err := minio.New(requiredEnv("S3_ENDPOINT"), &minio.Options{
		Creds: credentials.NewStaticV4(
			requiredEnv("S3_ACCESS_KEY"), requiredEnv("S3_SECRET_KEY"), ""),
		Secure: secure,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	const bucket = "otterio-sdk-example"
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
			log.Fatal(err)
		}
	}

	info, err := client.FPutObject(ctx, bucket, "hello.txt", "hello.txt",
		minio.PutObjectOptions{ContentType: "text/plain"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Uploaded %s (%d bytes)", info.Key, info.Size)

	object, err := client.GetObject(ctx, bucket, "hello.txt", minio.GetObjectOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer object.Close()
	if _, err := io.Copy(os.Stdout, object); err != nil {
		log.Fatal(err)
	}
}
```

在含有 `hello.txt` 的目录执行：

```sh
go run .
```

程序会记录上传对象的大小，并在读回对象后输出 `Hello from OtterIO SDK!`。`GetObject` 返回延迟读取的对象：错误可能出现在读取阶段，因此要检查 `io.Copy` 等读取操作的错误，并在使用结束后关闭对象。示例会将存储桶和上传的对象保留在测试服务中。

## API 与兼容性

- [API 参考](./docs/API.md)：本仓库维护的构造函数、选项、操作及示例片段，使用英文。
- [Go 包文档](https://pkg.go.dev/github.com/soulteary/otterio-sdk/v7)：已发布版本的公开类型和方法。
- [凭据提供器](./pkg/credentials)：静态凭据、环境变量、文件、IAM 和 STS。示例使用静态 Signature V4 凭据；临时凭据需要将会话令牌传入 `NewStaticV4` 的第三个参数。
- [加密](./pkg/encrypt)、[事件通知](./pkg/notification)、[生命周期](./pkg/lifecycle)和[标签](./pkg/tags)辅助包。

S3 操作能否使用，取决于目标服务的能力和配置。API 参考中标为 MinIO/AIStor 扩展的方法来自上游客户端，并不代表 OtterIO 服务端实现了这些功能。请结合 [OtterIO 服务端文档](https://github.com/soulteary/otterio/blob/main/README_zh_CN.md#了解更多)，在目标服务上验证应用实际使用的操作。AWS 特有操作也需要对应的 AWS 服务。

## 示例与开发

完整示例按操作分类列在[英文 README](./README.md#examples)。`examples/s3` 和 `examples/minio` 是独立的 Go 模块，每个示例文件都有自己的 `main` 函数。可用 `make examples` 编译所有示例；运行单个示例前，先修改端点、凭据、存储桶和本地文件路径：

```sh
cd examples/s3
go run listbuckets.go
```

这些示例目录中不能使用 `go run .` 或 `go run *.go`。多数示例保留了上游占位值或公共测试端点，需要先配置再运行；编译成功并不代表目标服务支持对应操作。

无需存储服务即可执行 SDK 的本地检查：

```sh
go test -short -race ./...
go build ./...
```

贡献流程、实时服务测试和其他检查见 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## 上游来源与许可证

本项目保留了上游的 `minio` Go 包名，以及部分 MinIO 特有 API 名称、协议字段和扩展说明。上游文档可作为背景资料；本分叉的行为请以本仓库的源代码和文档为准。

SDK 按 [Apache License, Version 2.0](./LICENSE) 分发，上游署名见 [NOTICE](./NOTICE)。“MinIO” 是 MinIO, Inc. 的商标，此处仅用于说明项目来源；Apache 许可证不授予商标权。
