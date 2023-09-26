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
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
)

func TestLogger(t *testing.T) {
	handler := Logger(func(ctx *core.Context, f func(...slog.Attr)) {
		f(slog.String("test", "test"))
	}).Handler(func(c *core.Context) {})

	rec := httptest.NewRecorder()
	c := core.AcquireContext(context.Background())
	c.ClientResponse = core.AcquireResponseWriter(rec)
	c.ClientRequest = httptest.NewRequest("GET", "/", nil)
	handler(c)
}
