/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2017 MinIO, Inc.
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

package credentials

import (
	"os"
	"path/filepath"
	"runtime"
)

// A FileOC retrieves credentials from the current user's home
// directory, and keeps track if those credentials are expired.
//
// Configuration file example: $HOME/.oc/config.json
type FileOC struct {
	// Path to the shared credentials file.
	//
	// If empty will look for "OC_SHARED_CREDENTIALS_FILE" env variable. If the
	// env value is empty will default to current user's home directory.
	// Linux/OSX: "$HOME/.oc/config.json"
	// Windows:   "%USERALIAS%\oc\config.json"
	Filename string

	// OC alias to extract credentials from the shared credentials file. If empty
	// will default to environment variable "OC_ALIAS" or "s3" if
	// environment variable is also not set.
	Alias string

	// retrieved states if the credentials have been successfully retrieved.
	retrieved bool
}

// NewFileOC returns a pointer to a new Credentials object
// wrapping the Alias file provider.
func NewFileOC(filename, alias string) *Credentials {
	return New(&FileOC{
		Filename: filename,
		Alias:    alias,
	})
}

func (p *FileOC) retrieve() (Value, error) {
	if p.Filename == "" {
		if value := os.Getenv("OC_SHARED_CREDENTIALS_FILE"); value != "" {
			p.Filename = value
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return Value{}, err
			}
			p.Filename = filepath.Join(homeDir, ".oc", "config.json")
			if runtime.GOOS == "windows" {
				p.Filename = filepath.Join(homeDir, "oc", "config.json")
			}
		}
	}

	if p.Alias == "" {
		p.Alias = os.Getenv("OC_ALIAS")
		if p.Alias == "" {
			p.Alias = "s3"
		}
	}

	p.retrieved = false

	hostCfg, err := loadAlias(p.Filename, p.Alias)
	if err != nil {
		return Value{}, err
	}

	p.retrieved = true
	return Value{
		AccessKeyID:     hostCfg.AccessKey,
		SecretAccessKey: hostCfg.SecretKey,
		SignerType:      parseSignatureType(hostCfg.API),
	}, nil
}

// Retrieve reads and extracts the shared credentials from the current
// users home directory.
func (p *FileOC) Retrieve() (Value, error) {
	return p.retrieve()
}

// RetrieveWithCredContext - is like Retrieve()
func (p *FileOC) RetrieveWithCredContext(_ *CredContext) (Value, error) {
	return p.retrieve()
}

// IsExpired returns if the shared credentials have expired.
func (p *FileOC) IsExpired() bool {
	return !p.retrieved
}
