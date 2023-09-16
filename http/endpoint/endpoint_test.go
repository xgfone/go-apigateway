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

package endpoint

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/internal/httpx"
	"github.com/xgfone/go-apigateway/http/upstream"
)

func TestNewEndpoint(t *testing.T) {
	ep := New("127.0.0.1", 80, 10)

	c := core.AcquireContext(context.Background())
	defer core.ReleaseContext(c)

	oldclient := upstream.DefaultHttpClient
	defer func() { upstream.DefaultHttpClient = oldclient }()
	upstream.DefaultHttpClient = &http.Client{
		Transport: httpx.RoundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 204,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}),
	}

	c.UpstreamRequest = &http.Request{URL: &url.URL{Path: "/"}}
	_resp, err := ep.Serve(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}

	resp, ok := _resp.(*http.Response)
	if !ok {
		t.Fatal("expect *http.Response, but got not")
	}

	if resp.StatusCode != 204 {
		t.Errorf("expect status code 204, but got %d", resp.StatusCode)
	}
	if c.UpstreamRequest.Host != ep.ID() {
		t.Errorf("expect host '%s', but got '%s'", ep.ID(), c.UpstreamRequest.Host)
	}
}
