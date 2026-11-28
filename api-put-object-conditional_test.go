// Copyright 2026 soulteary. Licensed under the Apache License, Version 2.0.

package minio

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/soulteary/otterio-sdk/v7/pkg/credentials"
)

// Exercise the public upload API against an HTTP endpoint: a condition on the
// initiation request alone cannot protect the object committed at completion.
func TestPutObjectConditionalHeaders(t *testing.T) {
	const objectSize = 3*1024 + 17
	for _, route := range []struct {
		name    string
		size    int64
		readAt  bool
		v2      bool
		direct  bool
		putOpts PutObjectOptions
	}{
		{name: "single PUT", size: objectSize, putOpts: PutObjectOptions{DisableMultipart: true}},
		{name: "multipart ReadAt", size: objectSize, readAt: true},
		{name: "multipart stream", size: objectSize},
		{name: "multipart unknown length", size: -1},
		{name: "multipart parallel", size: objectSize, putOpts: PutObjectOptions{ConcurrentStreamParts: true}},
		{name: "multipart parallel unknown length", size: -1, putOpts: PutObjectOptions{ConcurrentStreamParts: true}},
		{name: "multipart signature V2", size: objectSize, v2: true},
		{name: "direct completion", direct: true},
	} {
		t.Run(route.name, func(t *testing.T) {
			for _, condition := range []struct {
				name          string
				header        string
				etag          string
				want          string
				reject        bool
				embeddedError bool
			}{
				{name: "unconditional"},
				{name: "matching ETag", header: "If-Match", etag: "old-etag", want: `"old-etag"`},
				{name: "existing object", header: "If-Match", etag: "*", want: "*"},
				{name: "new object", header: "If-None-Match", etag: "*", want: "*"},
				{name: "different ETag", header: "If-None-Match", etag: "old-etag", want: `"old-etag"`},
				{name: "failed precondition", header: "If-Match", etag: "stale-etag", want: `"stale-etag"`, reject: true},
				{name: "failed precondition after HTTP 200", header: "If-Match", etag: "stale-etag", want: `"stale-etag"`, reject: true, embeddedError: true},
			} {
				// Only multipart completion can fail after keepalive whitespace
				// has already committed an HTTP 200 response.
				if condition.embeddedError && route.putOpts.DisableMultipart {
					continue
				}
				t.Run(condition.name, func(t *testing.T) {
					type request struct {
						phase  string
						header http.Header
					}
					var mu sync.Mutex
					var requests []request
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						query := r.URL.Query()
						var phase string
						switch {
						case r.Method == http.MethodPost && query.Has("uploads"):
							phase = "initiate"
						case r.Method == http.MethodPut && query.Get("partNumber") != "":
							phase = "part"
						case r.Method == http.MethodPost && query.Get("uploadId") != "":
							phase = "complete"
						case r.Method == http.MethodDelete && query.Get("uploadId") != "":
							phase = "abort"
						case r.Method == http.MethodPut:
							phase = "single"
						default:
							t.Errorf("unexpected request: %s %s", r.Method, r.URL)
							http.Error(w, "unexpected request", http.StatusBadRequest)
							return
						}
						mu.Lock()
						requests = append(requests, request{phase: phase, header: r.Header.Clone()})
						mu.Unlock()
						if _, err := io.Copy(io.Discard, r.Body); err != nil {
							t.Errorf("read %s body: %v", phase, err)
							return
						}

						w.Header().Set("Content-Type", "application/xml")
						if phase == "initiate" && (r.Header.Get("If-Match") != "" || r.Header.Get("If-None-Match") != "") {
							w.WriteHeader(http.StatusNotImplemented)
							io.WriteString(w, `<Error><Code>NotImplemented</Code><Message>Conditions require object completion</Message></Error>`)
							return
						}
						if condition.reject && (phase == "complete" || phase == "single") {
							if condition.embeddedError {
								w.WriteHeader(http.StatusOK)
								io.WriteString(w, xml.Header+" \t\r\n ")
								w.(http.Flusher).Flush()
							} else {
								w.WriteHeader(http.StatusPreconditionFailed)
							}
							io.WriteString(w, `<Error><Code>PreconditionFailed</Code><Message>ETag changed</Message></Error>`)
							return
						}
						switch phase {
						case "initiate":
							io.WriteString(w, `<InitiateMultipartUploadResult><Bucket>bucket</Bucket><Key>object</Key><UploadId>upload-id</UploadId></InitiateMultipartUploadResult>`)
						case "complete":
							io.WriteString(w, `<CompleteMultipartUploadResult><Bucket>bucket</Bucket><Key>object</Key><ETag>&quot;completed-etag-4&quot;</ETag></CompleteMultipartUploadResult>`)
						case "abort":
							w.WriteHeader(http.StatusNoContent)
						default:
							w.Header().Set("ETag", `"3858f62230ac3c915f300c664312c11f"`)
							w.WriteHeader(http.StatusOK)
						}
					}))
					t.Cleanup(server.Close)

					creds := credentials.NewStaticV4("test-access-key", "test-secret-key", "")
					if route.v2 {
						creds = credentials.NewStaticV2("test-access-key", "test-secret-key", "")
					}
					client, err := New(strings.TrimPrefix(server.URL, "http://"), &Options{
						Creds:        creds,
						Region:       "us-east-1",
						UploadLimits: &UploadLimits{MinPartSize: 1024, MaxPartSize: 1024, MaxPartsCount: 16},
					})
					if err != nil {
						t.Fatal(err)
					}
					opts := route.putOpts
					opts.PartSize = 1024
					opts.NumThreads = 2
					if !route.direct {
						opts.ContentType = "text/plain"
						opts.UserMetadata = map[string]string{"init-only": "metadata"}
						opts.customHeaders = http.Header{"X-Amz-Meta-Custom-Init-Only": {"custom"}}
					}
					switch condition.header {
					case "If-Match":
						opts.SetMatchETag(condition.etag)
					case "If-None-Match":
						opts.SetMatchETagExcept(condition.etag)
					}
					if route.direct {
						_, err = (Core{Client: client}).CompleteMultipartUpload(t.Context(), "bucket", "object", "upload-id",
							[]CompletePart{{PartNumber: 1, ETag: "3858f62230ac3c915f300c664312c11f"}}, opts)
					} else {
						var reader io.Reader = bytes.NewReader(bytes.Repeat([]byte("x"), objectSize))
						if !route.readAt {
							reader = struct{ io.Reader }{reader}
						}
						_, err = client.PutObject(t.Context(), "bucket", "object", reader, route.size, opts)
					}
					if condition.reject {
						errResponse := ToErrorResponse(err)
						if errResponse.Code != "PreconditionFailed" {
							t.Fatalf("error code = %q, want PreconditionFailed (error %v)", errResponse.Code, err)
						}
						wantStatus := http.StatusPreconditionFailed
						if condition.embeddedError {
							wantStatus = http.StatusOK
						}
						if errResponse.StatusCode != wantStatus {
							t.Errorf("error HTTP status = %d, want %d", errResponse.StatusCode, wantStatus)
						}
					} else if err != nil {
						t.Fatal(err)
					}

					mu.Lock()
					defer mu.Unlock()
					counts := make(map[string]int)
					for _, req := range requests {
						counts[req.phase]++
						if req.phase == "complete" || req.phase == "single" {
							for _, name := range []string{"If-Match", "If-None-Match"} {
								want := ""
								if name == condition.header {
									want = condition.want
								}
								if got := req.header.Get(name); got != want {
									t.Errorf("%s %s = %q, want %q", req.phase, name, got, want)
								}
							}
							if condition.header != "" && !route.v2 {
								_, authTail, found := strings.Cut(req.header.Get("Authorization"), "SignedHeaders=")
								signedHeaders, _, _ := strings.Cut(authTail, ",")
								if !found || !strings.Contains(";"+signedHeaders+";", ";"+strings.ToLower(condition.header)+";") {
									t.Errorf("%s condition is not bound to the SigV4 signature: %q", req.phase, signedHeaders)
								}
							}
						}
						if req.phase == "complete" {
							for _, name := range []string{"X-Amz-Meta-Init-Only", "X-Amz-Meta-Custom-Init-Only"} {
								if got := req.header.Get(name); got != "" {
									t.Errorf("initialization-only header %s leaked into completion: %q", name, got)
								}
							}
						}
						if req.phase == "initiate" || req.phase == "part" || req.phase == "abort" {
							if req.header.Get("If-Match") != "" || req.header.Get("If-None-Match") != "" {
								t.Errorf("destination condition leaked into %s request", req.phase)
							}
						}
						if req.phase == "initiate" {
							if req.header.Get("X-Amz-Meta-Init-Only") != "metadata" || req.header.Get("X-Amz-Meta-Custom-Init-Only") != "custom" {
								t.Error("initialization lost the object's metadata")
							}
						}
					}
					if got := counts["single"] + counts["complete"]; got != 1 {
						t.Errorf("commit requests = %d, want 1 (failed conditions must not be retried)", got)
					}
					if route.direct {
						if counts["complete"] != 1 || len(requests) != 1 {
							t.Errorf("direct completion request counts = %v", counts)
						}
					} else if opts.DisableMultipart {
						if counts["single"] != 1 || len(requests) != 1 {
							t.Errorf("single PUT request counts = %v", counts)
						}
					} else {
						if counts["initiate"] != 1 || counts["complete"] != 1 || counts["part"] < 2 {
							t.Errorf("automatic multipart request counts = %v", counts)
						}
						if condition.reject && counts["abort"] != 1 {
							t.Errorf("abort requests = %d, want 1 after rejected completion", counts["abort"])
						}
					}
				})
			}
		})
	}
}
