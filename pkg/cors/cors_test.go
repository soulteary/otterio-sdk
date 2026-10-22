/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2015-2024 MinIO, Inc.
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

package cors

import (
	"bytes"
	"os"
	"testing"
)

func TestCORSXMLMarshal(t *testing.T) {
	fileContents, err := os.ReadFile("testdata/example.xml")
	if err != nil {
		t.Fatal(err)
	}
	// Git may check this fixture out with CRLF on Windows. XML normalizes
	// line endings, so compare against the canonical LF representation and
	// exercise both input forms on every platform.
	canonical := bytes.ReplaceAll(fileContents, []byte("\r\n"), []byte("\n"))
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{name: "LF", data: canonical},
		{name: "CRLF", data: bytes.ReplaceAll(canonical, []byte("\n"), []byte("\r\n"))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := ParseBucketCorsConfig(bytes.NewReader(tc.data))
			if err != nil {
				t.Fatal(err)
			}
			remarshalled, err := c.ToXML()
			if err != nil {
				t.Fatal(err)
			}
			want := bytes.TrimSpace(canonical)
			if !bytes.Equal(want, remarshalled) {
				t.Errorf("got: %s, want: %s", remarshalled, want)
			}
		})
	}
}
