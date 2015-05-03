// +build ignore

package main

import (
	"fmt"

	"github.com/minio-io/objectstorage-go"
)

func main() {
	config := new(objectstorage.Config)
	config.Endpoint = "https://s3.amazonaws.com"
	config.AccessKeyID = "REDACTED_AWS_ACCESS_KEY_ID"
	config.SecretAccessKey = "REDACTED_AWS_SECRET_ACCESS_KEY"
	m := objectstorage.New(config)

	err := m.PutBucket("testbucket")
	fmt.Println(err)

	err = m.PutBucketACL("testbucket", "public-read")
	fmt.Println(err)

	err = m.PutBucketACL("testbucket", "invalid")
	fmt.Println(err)

	err = m.HeadBucket("testbucket")
	fmt.Println(err)

	_, err = m.ListBuckets()
	fmt.Println(err)
}
