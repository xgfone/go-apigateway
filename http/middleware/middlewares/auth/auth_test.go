// Copyright 2023 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
)

func TestAuthorization(t *testing.T) {
	mw, err := middleware.DefaultRegistry.Build("auth", map[string]any{
		"type": "token",
		"auths": map[string]map[string]string{
			"user1": map[string]string{"x-user-id": "1"},
			"user2": map[string]string{"x-user-id": "2"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	handler := mw.Handler(func(c *core.Context) {})
	c := core.AcquireContext(context.Background())
	c.ClientRequest = &http.Request{Header: make(http.Header, 2)}

	c.Error = nil
	c.IsAborted = false
	handler(c)
	if err := errors.Unwrap(c.Error); err != errMissingAuthorization {
		t.Errorf("expect error '%v', but got '%v'", errMissingAuthorization, err)
	}

	c.Error = nil
	c.IsAborted = false
	c.ClientRequest.Header.Set("Authorization", "user1")
	handler(c)
	if err := errors.Unwrap(c.Error); err != errMissingAuthType {
		t.Errorf("expect error '%v', but got '%v'", errMissingAuthType, err)
	}

	c.Error = nil
	c.IsAborted = false
	c.ClientRequest.Header.Set("Authorization", "basic user1")
	handler(c)
	if err := errors.Unwrap(c.Error); err != errInvalidAuthType {
		t.Errorf("expect error '%v', but got '%v'", errInvalidAuthType, err)
	}

	c.Error = nil
	c.IsAborted = false
	c.ClientRequest.Header.Set("Authorization", "token user")
	handler(c)
	if err := errors.Unwrap(c.Error); err != errInvalidAuth {
		t.Errorf("expect error '%v', but got '%v'", errInvalidAuth, err)
	}

	c.Error = nil
	c.IsAborted = false
	c.ClientRequest.Header.Set("Authorization", "token user1")
	handler(c)
	if c.Error != nil {
		t.Errorf("unexpect any error, but got '%v'", c.Error)
	} else if id := c.ClientRequest.Header.Get("X-User-Id"); id != "1" {
		t.Errorf("expect user id '%s', but got '%s'", "1", id)
	}
}
