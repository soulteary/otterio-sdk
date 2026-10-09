//go:build example
// +build example

/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2026 MinIO, Inc.
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

	// Note: my-bucketname, my-objectname and my-annotationname are dummy
	// values, please replace them with original values.

	s3Client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, os.Getenv("S3_SESSION_TOKEN")),
		Secure: secure,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Annotation deletion is permanent and irreversible.
	err = s3Client.RemoveObjectAnnotation(context.Background(), "my-bucketname", "my-objectname", "model.labels.json", minio.RemoveObjectAnnotationOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("annotation removed")
}
