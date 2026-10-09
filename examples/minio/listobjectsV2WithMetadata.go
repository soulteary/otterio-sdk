//go:build example
// +build example

/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2019 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/soulteary/otterio-sdk/v7"
	"github.com/soulteary/otterio-sdk/v7/pkg/credentials"
)

func main() {
	requiredEnv := func(name string) string {
		value := os.Getenv(name)
		if value == "" {
			log.Fatalf("Set %s before running this example", name)
		}
		return value
	}
	endpoint := requiredEnv("S3_ENDPOINT")
	accessKey := requiredEnv("S3_ACCESS_KEY")
	secretKey := requiredEnv("S3_SECRET_KEY")
	secure := true
	if value := os.Getenv("S3_USE_TLS"); value != "" {
		var err error
		secure, err = strconv.ParseBool(value)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Note: YOUR-ACCESSKEYID, YOUR-SECRETACCESSKEY, my-bucketname and my-prefixname
	// are dummy values, please replace them with original values.

	// Requests are always secure (HTTPS) by default. Set secure=false to enable insecure (HTTP) access.
	// This boolean value is the last argument for New().

	// New returns an Amazon S3 compatible client object. API compatibility (v2 or v4) is automatically
	// determined based on the Endpoint value.
	s3Client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, os.Getenv("S3_SESSION_TOKEN")),
		Secure: secure,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	opts := minio.ListObjectsOptions{
		WithMetadata: true,
		Prefix:       "my-prefixname",
		Recursive:    true,
	}

	// List all objects from a bucket-name with a matching prefix.
	for object := range s3Client.ListObjects(context.Background(), "my-bucketname", opts) {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}
		fmt.Println(object)
	}
	return
}
