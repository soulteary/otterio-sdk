<div align="center">

[![OtterIO Go SDK — Go Client for S3-Compatible Object Storage](./.github/otterio-sdk-banner.png)](https://github.com/soulteary/otterio-sdk)

# OtterIO Go SDK

**Go Client for S3-Compatible Object Storage**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.27.2%2B-00ADD8.svg?logo=go&logoColor=white)](./go.mod)
[![GitHub](https://img.shields.io/badge/GitHub-soulteary%2Fotterio--sdk-181717.svg?logo=github)](https://github.com/soulteary/otterio-sdk)

English · [简体中文](./README_zh_CN.md)

</div>

OtterIO SDK is a Go client for Amazon S3 compatible object storage. Connect Go applications to [OtterIO](https://github.com/soulteary/otterio) or other S3-compatible services to manage buckets, upload and download objects, and generate presigned URLs.

This README covers [installation](#install) and a complete [upload and read-back example](#quick-start-upload-and-read-an-object). The [API reference](./docs/API.md) and [examples](#examples) cover more operations. [OtterIO](https://github.com/soulteary/otterio) provides the storage server, [OC](https://github.com/soulteary/oc) provides command-line operations and administration, and this repository provides the Go client library.

> [!IMPORTANT]
> OtterIO SDK is an independently maintained fork of [MinIO Go SDK](https://github.com/minio/minio-go). It is **not** affiliated with, endorsed by, or sponsored by MinIO, Inc. See [Upstream and license](#upstream-and-license) for attribution and licensing; original copyright notices are retained in [NOTICE](./NOTICE) and the source files.

---

## What is OtterIO SDK

The SDK provides bucket and object operations, presigned URLs, and helper packages for credentials, encryption, notifications, lifecycle rules, and tags. Availability of each operation depends on the target service's features and configuration; see [API reference and compatibility](#api-reference-and-compatibility).

The module path is `github.com/soulteary/otterio-sdk/v7`. The Go package name remains `minio`, so examples use an explicit `minio` import alias. When migrating an application, update imports for both the root module and helper packages such as `pkg/credentials` and `pkg/encrypt`.

---

## Install

Use Go **1.27.2 or newer**, as declared in [go.mod](./go.mod), in a Go module:

```sh
mkdir otterio-sdk-example
cd otterio-sdk-example
go mod init example.com/otterio-sdk-example
go get github.com/soulteary/otterio-sdk/v7
```

For an existing module, run only the `go get` command. To select a release, append its tag to the module path, for example `@v7.3.1`; see the repository's [tags](https://github.com/soulteary/otterio-sdk/tags).

---

## Quick start: upload and read an object

Start an OtterIO server using its [quick start](https://github.com/soulteary/otterio#quick-start), or use an existing S3 service. The example needs credentials allowed to check/create a bucket, upload an object, and read it. It creates `otterio-sdk-example` if needed and uploads `hello.txt`; running it again writes the same object key. No command-line storage client is required.

In the same shell as your local OtterIO setup, reuse the credentials used to start the server:

```sh
: "${OTTERIO_ROOT_USER:?Set the username used to start OtterIO}"
: "${OTTERIO_ROOT_PASSWORD:?Set the password used to start OtterIO}"
export S3_ENDPOINT=127.0.0.1:9000
export S3_ACCESS_KEY="$OTTERIO_ROOT_USER"
export S3_SECRET_KEY="$OTTERIO_ROOT_PASSWORD"
export S3_USE_TLS=false
printf 'Hello from OtterIO SDK!\n' > hello.txt
```

For another service, set `S3_ENDPOINT`, `S3_ACCESS_KEY`, and `S3_SECRET_KEY` to its connection details. Use the **S3 API port**, not the web console port. `S3_ENDPOINT` can be a `host:port` without a URL path; `S3_USE_TLS=true` selects HTTPS and is the example's default. An explicit `http://` or `https://` scheme is also accepted if it matches `Secure`. Use `false` for the loopback HTTP server above, and use HTTPS for remote connections. Adjust the bucket name and `MakeBucketOptions.Region` for your service.

Save the following as `main.go` in the module directory:

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

Run it from the directory containing `hello.txt`:

```sh
go run .
```

The program logs the uploaded object size and prints `Hello from OtterIO SDK!` after reading the object back. `GetObject` returns a lazy reader: errors can occur when reading, so check `io.Copy` (or other read operations) and close the object when done. The sample leaves the bucket and uploaded object in your test server.

---

## API reference and compatibility

- [API reference](./docs/API.md): constructors, options, operations, and snippets maintained in this repository.
- [Go package documentation](https://pkg.go.dev/github.com/soulteary/otterio-sdk/v7): exported Go types and methods for published versions.
- [Credentials](./pkg/credentials): static, environment, file, IAM, and STS providers. The example uses static Signature V4 credentials; supply the session token as the third argument to `NewStaticV4` for temporary credentials.
- [Encryption helpers](./pkg/encrypt), [notifications](./pkg/notification), [lifecycle](./pkg/lifecycle), and [tags](./pkg/tags).

S3 operations depend on the target server's features and configuration. Methods described as MinIO/AIStor extensions in the API reference are inherited client APIs and do not establish OtterIO server support for those features. Check the [OtterIO server documentation](https://github.com/soulteary/otterio#further-reading) and validate the operations your application uses against your target server. AWS-specific operations likewise require the corresponding AWS service.

---

## Examples

Examples that previously used an upstream demo service now require `S3_ENDPOINT`,
`S3_ACCESS_KEY`, and `S3_SECRET_KEY`. Set `S3_USE_TLS=false` only for your local HTTP
server; TLS is the default. Extension examples require support from the target server.

### Bucket Operations

-	[makebucket.go](./examples/s3/makebucket.go)
-	[listbuckets.go](./examples/s3/listbuckets.go)
-	[bucketexists.go](./examples/s3/bucketexists.go)
-	[removebucket.go](./examples/s3/removebucket.go)
-	[listobjects.go](./examples/s3/listobjects.go)
-	[listobjectsV2.go](./examples/s3/listobjectsV2.go)
-	[listincompleteuploads.go](./examples/s3/listincompleteuploads.go)

### Bucket policy Operations

-	[setbucketpolicy.go](./examples/s3/setbucketpolicy.go)
-	[getbucketpolicy.go](./examples/s3/getbucketpolicy.go)

### Bucket lifecycle Operations

-	[setbucketlifecycle.go](./examples/s3/setbucketlifecycle.go)
-	[getbucketlifecycle.go](./examples/s3/getbucketlifecycle.go)

### Bucket encryption Operations

-	[setbucketencryption.go](./examples/s3/setbucketencryption.go)
-	[getbucketencryption.go](./examples/s3/getbucketencryption.go)
-	[removebucketencryption.go](./examples/s3/removebucketencryption.go)

### Bucket replication Operations

-	[setbucketreplication.go](./examples/s3/setbucketreplication.go)
-	[getbucketreplication.go](./examples/s3/getbucketreplication.go)
-	[removebucketreplication.go](./examples/s3/removebucketreplication.go)

### Bucket notification Operations

-	[setbucketnotification.go](./examples/s3/setbucketnotification.go)
-	[getbucketnotification.go](./examples/s3/getbucketnotification.go)
-	[removeallbucketnotification.go](./examples/s3/removeallbucketnotification.go)
-	[listenbucketnotification.go](./examples/minio/listenbucketnotification.go) (MinIO Extension)
-	[listen-notification.go](./examples/minio/listen-notification.go) (MinIO Extension)

### File Object Operations

-	[fputobject.go](./examples/s3/fputobject.go)
-	[fgetobject.go](./examples/s3/fgetobject.go)

### Object Operations

-	[putobject.go](./examples/s3/putobject.go)
-	[getobject.go](./examples/s3/getobject.go)
-	[statobject.go](./examples/s3/statobject.go)
-	[copyobject.go](./examples/s3/copyobject.go)
-	[removeobject.go](./examples/s3/removeobject.go)
-	[removeincompleteupload.go](./examples/s3/removeincompleteupload.go)
-	[removeobjects.go](./examples/s3/removeobjects.go)
-	[putobjectannotation.go](./examples/s3/putobjectannotation.go)
-	[getobjectannotation.go](./examples/s3/getobjectannotation.go)
-	[listobjectannotations.go](./examples/s3/listobjectannotations.go)
-	[removeobjectannotation.go](./examples/s3/removeobjectannotation.go)

### Encrypted Object Operations

-	[put-encrypted-object.go](./examples/s3/put-encrypted-object.go)
-	[get-encrypted-object.go](./examples/s3/get-encrypted-object.go)
-	[fputencrypted-object.go](./examples/s3/fputencrypted-object.go)

### Presigned Operations

-	[presignedgetobject.go](./examples/s3/presignedgetobject.go)
-	[presignedputobject.go](./examples/s3/presignedputobject.go)
-	[presignedheadobject.go](./examples/s3/presignedheadobject.go)
-	[presignedpostpolicy.go](./examples/s3/presignedpostpolicy.go)

---

## Development and contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for local checks and live-server tests. To check the SDK without a storage server:

```sh
go test -short -race ./...
go build ./...
```

The `examples/s3` and `examples/minio` directories are separate Go modules. They contain independent programs with their own `main` functions. Build them with `make examples`, or run one file from its own directory after editing its endpoint, credentials, bucket, and local file paths:

```sh
cd examples/s3
go run listbuckets.go
```

Do not use `go run .` or `go run *.go` inside these example directories. Most examples retain upstream placeholders or public test endpoints and need configuration before use. Compiling them does not validate server compatibility.

---

## Upstream and license

This fork retains the upstream `minio` Go package name and some MinIO-specific API names, protocol fields, and extension descriptions. Upstream documentation is useful background, but the source and documentation in this repository describe this fork.

The SDK is distributed under the [Apache License, Version 2.0](./LICENSE). See [NOTICE](./NOTICE) for upstream attribution. “MinIO” is a trademark of MinIO, Inc., used here to identify the upstream project; the Apache license does not grant trademark rights.
