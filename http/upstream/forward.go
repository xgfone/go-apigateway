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

package upstream

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/upstream"
	"github.com/xgfone/go-defaults"
)

// Forward forwards the http request by the upstream.
func Forward(c *core.Context) {
	if c.IsAborted {
		return
	}

	up, ok := upstream.Manager.Get(c.UpstreamId)
	if !ok {
		c.Abort(fmt.Errorf("no upstream '%s'", c.UpstreamId))
		return
	}

	c.Upstream = up
	if c.UpstreamRequest == nil {
		c.UpstreamRequest = newRequest(c)
	}

	// Set the url scheme.
	switch scheme := up.Scheme(); scheme {
	case "https", "http":
		c.UpstreamRequest.URL.Scheme = scheme
	default:
		if c.UpstreamRequest.URL.Scheme == "" {
			c.UpstreamRequest.URL.Scheme = "http"
		}
	}

	// Set the Host header.
	switch host := up.Host(); host {
	case "", HostClient:
		c.UpstreamRequest.Host = c.ClientRequest.Host
	case HostServer:
		c.UpstreamRequest.Host = "" // Clear the host and let the endpoint reset it.
	default:
		c.UpstreamRequest.Host = host
	}

	// Reset the url path
	if path := up.Path(); path != "" {
		c.UpstreamRequest.URL.Path = path
	}

	c.CallbackOnForward()
	if c.IsAborted {
		return
	}

	start := time.Now()

	var resp interface{}
	resp, c.Error = up.Serve(c.UpstreamRequest.Context(), c)
	if resp != nil {
		c.UpstreamResponse = resp.(*http.Response)
	}

	_log(c, up.Balancer().Policy(), time.Since(start), c.Error)
}

func _log(c *core.Context, policy string, cost time.Duration, err error) {
	req := c.UpstreamRequest
	if err != nil {
		slog.Error("fail to forward the http request",
			slog.String("reqid", defaults.GetRequestID(c.Context, req)),
			slog.String("upstream", c.UpstreamId),
			slog.String("balancer", policy),
			slog.String("method", req.Method),
			slog.String("hostname", req.Host),
			slog.String("host", req.URL.Host),
			slog.String("path", req.URL.Path),
			slog.String("query", req.URL.RawQuery),
			slog.String("cost", cost.String()),
			slog.Any("header", req.Header),
			slog.String("err", err.Error()),
		)
	} else if slog.Default().Enabled(c.Context, slog.LevelDebug) {
		slog.Debug("forward the http request",
			slog.String("reqid", defaults.GetRequestID(c.Context, req)),
			slog.String("upstream", c.UpstreamId),
			slog.String("balancer", policy),
			slog.String("method", req.Method),
			slog.String("hostname", req.Host),
			slog.String("host", req.URL.Host),
			slog.String("path", req.URL.Path),
			slog.String("query", req.URL.RawQuery),
			slog.String("cost", cost.String()),
			slog.Any("header", req.Header),
		)
	}
}
