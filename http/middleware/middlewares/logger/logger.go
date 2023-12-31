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

// Package provides a logger middleware to log the request.
package logger

import (
	"log/slog"
	"sync"
	"time"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
	"github.com/xgfone/go-apigateway/http/statuscode"
)

var (
	// Collect is used to collect the extra key-value attributes if set.
	//
	// Default: nil
	Collect func(c *core.Context, append func(...slog.Attr))

	// Enabled is used to decide whether to log the request if set.
	//
	// Default: nil
	Enabled func(c *core.Context) bool
)

// Logger returns a logger middleware to log the http request.
func Logger() middleware.Middleware { return logger }

var logger = middleware.New("logger", nil, func(next core.Handler) core.Handler {
	return func(c *core.Context) { logreq(c, next) }
})

func logreq(c *core.Context, next core.Handler) {
	if (Enabled != nil && !Enabled(c)) || !slog.Default().Enabled(c.Context, slog.LevelInfo) {
		next(c)
		return
	}

	start := time.Now()
	next(c)
	cost := time.Since(start)

	req := c.ClientRequest
	logattrs := getattrs()
	logattrs.Append(
		slog.String("reqid", c.RequestID()),
		slog.String("raddr", req.RemoteAddr),
		slog.String("method", req.Method),
		slog.String("host", req.Host),
		slog.String("path", req.URL.Path),
	)

	if c.RouteId != "" {
		logattrs.Append(slog.String("route", c.RouteId))
		if c.UpstreamId != "" {
			logattrs.Append(slog.String("upstream", c.UpstreamId))
		}
		if c.Endpoint != nil {
			logattrs.Append(slog.String("endpoint", c.Endpoint.ID()))
		}
	}

	logattrs.Append(
		slog.Int("code", c.ClientResponse.StatusCode()),
		slog.String("cost", cost.String()),
	)

	if Collect != nil {
		Collect(c, logattrs.Append)
	}

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
