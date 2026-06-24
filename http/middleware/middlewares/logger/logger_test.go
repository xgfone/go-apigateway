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

package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-loadbalancer/endpoint"
)

func TestLogger(t *testing.T) {
	origLogger := slog.Default()
	defer func() { slog.SetDefault(origLogger) }()

	buf := bytes.NewBuffer(nil)
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))

	config := NewDefaultConfig()
	config.LogExtra = func(_ *core.Context, append func(...slog.Attr)) {
		append(slog.String("test", "true"))
	}
	handler := config.Logger().Handler(func(c *core.Context) {})

	rec := httptest.NewRecorder()
	c := core.AcquireContext(context.Background())
	c.ClientRequest = httptest.NewRequest("GET", "/path", nil)
	c.ClientResponse = core.AcquireResponseWriter(rec)
	defer core.ReleaseResponseWriter(c.ClientResponse)

	c.RouteId = "routeid"
	c.UpstreamId = "upid"
	c.Endpoint = endpoint.New("epid", nil)
	handler(c)

	type Log struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
		Path  string `json:"path"`
		Test  string `json:"test"`

		RouteId    string `json:"route"`
		UpstreamId string `json:"upstream"`
	}

	expect := Log{
		Level: "INFO",
		Msg:   "log http request",
		Path:  "/path",
		Test:  "true",

		RouteId:    "routeid",
		UpstreamId: "upid",
	}

	var actual Log
	if err := json.Unmarshal(buf.Bytes(), &actual); err != nil {
		t.Fatal(err)
	} else if actual != expect {
		t.Errorf("expect %+v, but got %+v", expect, actual)
	}
}
