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

package router

import (
	"log/slog"
	"sync"
	"time"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/statuscode"
)

// DisableLog sets whether the router disables the logging or not.
func (r *Router) DisableLog(disable bool) { r.notlog.Store(disable) }

func (r *Router) log(c *core.Context, start time.Time, matched bool) {
	if r.notlog.Load() {
		return
	}

	req := c.ClientRequest
	cost := time.Since(start)
	logattrs := getattrs()
	logattrs.Append(
		slog.String("reqid", c.RequestID()),
		slog.String("raddr", req.RemoteAddr),
		slog.String("method", req.Method),
		slog.String("path", req.URL.Path),
		slog.String("query", req.URL.RawQuery),
	)

	if matched {
		logattrs.Append(slog.String("route", c.RouteId))
		if c.UpstreamId != "" {
			logattrs.Append(slog.String("upstream", c.UpstreamId))
		}
		if c.Endpoint != nil {
			logattrs.Append(slog.String("endpoint", c.Endpoint.ID()))
		}
	}

	logattrs.Append(
		slog.String("cost", cost.String()),
		slog.Int("code", c.ClientResponse.StatusCode()),
	)

	switch e := c.Error.(type) {
	case nil:
	case statuscode.Error:
		if e.Err != nil {
			slog.Any("err", e.Err)
		}
	default:
		logattrs.Append(slog.Any("err", c.Error))
	}

	slog.LogAttrs(c.Context, slog.LevelInfo, "log http request", logattrs.Attrs...)
	putattrs(logattrs)
}

type logattr struct{ Attrs []slog.Attr }

func (l *logattr) Append(attrs ...slog.Attr) { l.Attrs = append(l.Attrs, attrs...) }

var attrspool = &sync.Pool{New: func() any { return &logattr{make([]slog.Attr, 0, 12)} }}

func getattrs() *logattr  { return attrspool.Get().(*logattr) }
func putattrs(a *logattr) { clear(a.Attrs); a.Attrs = a.Attrs[:0]; attrspool.Put(a) }
