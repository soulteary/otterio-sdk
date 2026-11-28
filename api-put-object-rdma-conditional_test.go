// Copyright 2026 soulteary. Licensed under the Apache License, Version 2.0.

package minio

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"unsafe"
)

type rdmaConditionalTransport struct {
	requests int
}

func (rt *rdmaConditionalTransport) RoundTrip(*http.Request) (*http.Response, error) {
	rt.requests++
	return nil, errors.New("unexpected HTTP fallback from conditional RDMA upload")
}

// Both public RDMA dispatch paths must reject conditions before the native
// library can perform an unconditional write. This runs in ordinary stub
// builds and in -tags=rdma builds without requiring a device or native call.
func TestPutObjectRDMARejectsConditionalWrites(t *testing.T) {
	match, except := PutObjectOptions{}, PutObjectOptions{}
	match.SetMatchETag("old-etag")
	except.SetMatchETagExcept("*")
	for _, condition := range []struct {
		name    string
		headers http.Header
	}{
		{name: "matching ETag", headers: match.customHeaders},
		{name: "new object", headers: except.customHeaders},
		{name: "existing object", headers: http.Header{"If-Match": {"*"}}},
		{name: "different ETag", headers: http.Header{"If-None-Match": {`"old-etag"`}}},
		{name: "empty match", headers: http.Header{"If-Match": {""}}},
		{name: "empty except", headers: http.Header{"If-None-Match": {""}}},
		{name: "nil match", headers: http.Header{"If-Match": nil}},
		{name: "nil except", headers: http.Header{"If-None-Match": nil}},
		{name: "lowercase match", headers: http.Header{"if-match": {`"old-etag"`}}},
		{name: "lowercase except", headers: http.Header{"if-none-match": {"*"}}},
	} {
		for _, mode := range []string{"buffer", "stream"} {
			t.Run(condition.name+"/"+mode, func(t *testing.T) {
				transport := &rdmaConditionalTransport{}
				client, err := New("localhost:9000", &Options{
					EnableRDMA: true,
					Region:     "us-east-1",
					Transport:  transport,
				})
				if err != nil {
					t.Fatal(err)
				}
				// Even if the condition guard regresses in an RDMA build, fail
				// safely before native initialization or transfer can occur.
				client.rdmaOnce.Do(func() {
					client.rdmaInitErr = errors.New("unexpected native RDMA upload")
				})
				opts := PutObjectOptions{
					RDMABuffer:     unsafe.Pointer(new(byte)),
					RDMABufferSize: 1,
					customHeaders:  condition.headers,
				}
				source := strings.NewReader("x")
				var reader io.Reader
				if mode == "stream" {
					reader = source
				}
				info, err := client.PutObject(t.Context(), "bucket", "object", reader, 1, opts)
				errResponse := ToErrorResponse(err)
				if errResponse.Code != InvalidArgument || errResponse.StatusCode != http.StatusBadRequest || !strings.Contains(errResponse.Message, "RDMA") {
					t.Fatalf("expected conditional RDMA rejection, got %+v (error %v)", errResponse, err)
				}
				if info != (UploadInfo{}) {
					t.Errorf("rejected upload returned success info: %+v", info)
				}
				if transport.requests != 0 || source.Len() != 1 {
					t.Errorf("rejected upload made %d HTTP requests and consumed %d bytes", transport.requests, 1-source.Len())
				}
			})
		}
	}
}
