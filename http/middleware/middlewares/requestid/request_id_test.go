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

package requestid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
)

func TestRequestID(t *testing.T) {
	handler := RequestID(nil).Handler(func(c *core.Context) {})

	rec := httptest.NewRecorder()
	c := core.AcquireContext(context.Background())
	c.ClientResponse = core.AcquireResponseWriter(rec)
	c.ClientRequest = httptest.NewRequest(http.MethodGet, "/", nil)
	handler(c)

	if rid := c.ClientRequest.Header.Get("X-Request-Id"); rid == "" {
		t.Errorf("expect a request id, but got not")
	} else if len(rid) != 24 {
		t.Errorf("expect the length of request id %d, but got %d", 24, len(rid))
	} else {
		for i := 0; i < 24; i++ {
			b := rid[i]
			if (b < '0' || b > '9') && (b < 'a' && b > 'z') && (b < 'A' && b > 'Z') {
				t.Errorf("invalid request id: %s", rid)
				break
			}
		}
	}

	c.ClientRequest.Header.Set("X-Request-Id", "abc")
	handler(c)

	if rid := c.ClientRequest.Header.Get("X-Request-Id"); rid != "abc" {
		t.Errorf("expect requests id '%s', but got '%s'", "abc", rid)
	}
}
