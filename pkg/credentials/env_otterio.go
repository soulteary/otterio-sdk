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

import "os"

// A EnvOtterIO retrieves credentials from the environment variables of the
// running process. Environment credentials never expire.
//
// Environment variables used:
//
// * Access Key ID:     OTTERIO_ACCESS_KEY.
// * Secret Access Key: OTTERIO_SECRET_KEY.
// * Access Key ID:     OTTERIO_ROOT_USER.
// * Secret Access Key: OTTERIO_ROOT_PASSWORD.
type EnvOtterIO struct {
	retrieved bool
}

// NewEnvOtterIO returns a pointer to a new Credentials object
// wrapping the environment variable provider.
func NewEnvOtterIO() *Credentials {
	return New(&EnvOtterIO{})
}

func (e *EnvOtterIO) retrieve() (Value, error) {
	e.retrieved = false

	id := os.Getenv("OTTERIO_ROOT_USER")
	secret := os.Getenv("OTTERIO_ROOT_PASSWORD")

	signerType := SignatureV4
	if id == "" || secret == "" {
		id = os.Getenv("OTTERIO_ACCESS_KEY")
		secret = os.Getenv("OTTERIO_SECRET_KEY")
		if id == "" || secret == "" {
			signerType = SignatureAnonymous
		}
	}

	e.retrieved = true
	return Value{
		AccessKeyID:     id,
		SecretAccessKey: secret,
		SignerType:      signerType,
	}, nil
}

// Retrieve retrieves the keys from the environment.
func (e *EnvOtterIO) Retrieve() (Value, error) {
	return e.retrieve()
}

// RetrieveWithCredContext is like Retrieve() (no-op input cred context)
func (e *EnvOtterIO) RetrieveWithCredContext(_ *CredContext) (Value, error) {
	return e.retrieve()
}

// IsExpired returns if the credentials have been retrieved.
func (e *EnvOtterIO) IsExpired() bool {
	return !e.retrieved
}
