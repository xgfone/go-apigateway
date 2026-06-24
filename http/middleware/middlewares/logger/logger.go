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
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/statuscode"
)

// Config configures a request logger middleware handler.
type Config struct {
	// Enabled reports whether the request should be logged.
	//
	// If nil, all requests are logged when the configured slog level is enabled.
	Enabled func(*core.Context) bool

	// GetRequestBody returns the request body written to the "reqbody"
	// log attribute.
	//
	// If nil, the request body is omitted.
	GetRequestBody func(*core.Context) any

	// GetResponseBody returns the response body written to the "resbody",
	//
	// If nil, the response body is omitted.
	GetResponseBody func(*core.Context) any

	// PostHandler is called after PreHandle is called and before the logger
	// middleware handler ends.
	//
	// If nil, do nothing.
	PostHandle func(*core.Context)

	// PreHandler is called immediately before the next handler.
	//
	// If nil, do nothing.
	PreHandle func(*core.Context)

	// LogExtra appends the extra log attributes.
	//
	// If nil, do nothing.
	LogExtra func(c *core.Context, append func(...slog.Attr))
}

// Logger return a new Logger.
func (c Config) Logger() *Logger {
	logger := &Logger{
		enabled:         c.Enabled,
		getRequestBody:  c.GetRequestBody,
		getResponseBody: c.GetResponseBody,
		postHandle:      c.PostHandle,
		preHandle:       c.PreHandle,
		logExtra:        c.LogExtra,
		level:           slog.LevelInfo,
	}
	return logger
}

// NewDefaultConfig returns a new default Config.
//
//	Enabled: only ignore the root path "/".
func NewDefaultConfig() Config {
	return Config{
		Enabled: enabled,
	}
}

func enabled(c *core.Context) bool {
	return c.ClientRequest.URL.Path != "/"
}

// Logger is a logger middleware handler.
type Logger struct {
	enabled         func(*core.Context) bool
	getRequestBody  func(*core.Context) any
	getResponseBody func(*core.Context) any
	postHandle      func(*core.Context)
	preHandle       func(*core.Context)
	logExtra        func(*core.Context, func(...slog.Attr))

	level slog.Level
	next  core.Handler
}

// Name implements the interface middleware.Middleware#Name.
func (l *Logger) Name() string {
	return "logger"
}

// Handler implements the interface middleware.Middleware#Handler.
func (l *Logger) Handler(next core.Handler) core.Handler {
	if next == nil {
		panic("Logger.Handler: next core.Handler is nil")
	}

	_l := *l
	_l.next = next
	return _l.Serve
}

func enableLevel(ctx context.Context, level slog.Level) bool {
	return slog.Default().Enabled(ctx, level)
}

// Serve implements core.Handler.
func (l *Logger) Serve(c *core.Context) {
	if l.next == nil {
		c.Abort(statuscode.ErrInternalServerError.WithMessage("Logger: NO NEXT HANDLER"))
		return
	}

	if !(enableLevel(c.Context, l.level) && (l.enabled == nil || l.enabled(c))) {
		l.next(c)
		return
	}

	if l.preHandle != nil {
		l.preHandle(c)
	}
	if l.postHandle != nil {
		defer l.postHandle(c)
	}

	start := time.Now()
	l.next(c)
	cost := time.Since(start)

	req := c.ClientRequest
	logattrs := getattrs()
	logattrs.Append(
		slog.String("reqid", c.RequestID()),
		slog.String("raddr", req.RemoteAddr),
		slog.String("method", req.Method),
		slog.String("host", req.Host),
		slog.String("path", req.URL.Path),
		slog.String("query", req.URL.RawQuery),
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

	if l.logExtra != nil {
		l.logExtra(c, logattrs.Append)
	}

	logattrs.Append(
		slog.String("cost", cost.String()),
		slog.Int("code", c.ClientResponse.StatusCode()),
		slog.Any("reqheader", c.ClientRequest.Header),
		slog.Any("resheader", c.ClientResponse.Header()),
	)

	if l.getRequestBody != nil {
		logattrs.Append(slog.Any("reqbody", l.getRequestBody(c)))
	}
	if l.getResponseBody != nil {
		logattrs.Append(slog.Any("resbody", l.getResponseBody(c)))
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

	slog.LogAttrs(c.Context, l.level, "log http request", logattrs.Attrs...)
	putattrs(logattrs)
}

type logattr struct{ Attrs []slog.Attr }

func (l *logattr) Append(attrs ...slog.Attr) { l.Attrs = append(l.Attrs, attrs...) }

var attrspool = &sync.Pool{New: func() any { return &logattr{make([]slog.Attr, 0, 12)} }}

func getattrs() *logattr  { return attrspool.Get().(*logattr) }
func putattrs(a *logattr) { clear(a.Attrs); a.Attrs = a.Attrs[:0]; attrspool.Put(a) }
