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

package runtime

import (
	"context"
	"log/slog"
	"net/http"
	"net/netip"
	"net/url"
	"sync"

	"github.com/xgfone/go-apigateway/pkg/internal/httpx"
	"github.com/xgfone/go-apigateway/pkg/nets"
	"github.com/xgfone/go-defaults"
	"github.com/xgfone/go-defaults/assists"
	"github.com/xgfone/go-loadbalancer"
)

// NewResponseWriter is used to customize the wrapper
// from http.ResponseWriter to ResponseWriter.
var NewResponseWriter func(http.ResponseWriter) ResponseWriter = newResponseWriter

func newResponseWriter(w http.ResponseWriter) ResponseWriter {
	return httpx.NewResponseWriter(w)
}

var ctxpool = &sync.Pool{New: func() any {
	return &Context{
		forwards:    make([]func(), 0, 2),
		respbodys:   make([]func(), 0, 2),
		respheaders: make([]func(), 0, 4),
	}
}}

// AcquireContext acquires a context from the pool.
func AcquireContext(ctx context.Context) *Context {
	c := ctxpool.Get().(*Context)
	c.Context = ctx
	return c
}

// ReleaseContext releases the context to the pool.
func ReleaseContext(c *Context) { c.Reset(); ctxpool.Put(c) }

// Handler is used to handle the http request.
type Handler func(c *Context)

// ResponseWriter is the extension of http.ResponseWriter.
type ResponseWriter interface {
	http.ResponseWriter
	WroteHeader() bool
	StatusCode() int
	Written() int
}

// Context represents a request context.
type Context struct {
	// Mode        // Set by the router after matching the route or caller when calling the route.

	Route          Route           // Set by the router after matching the route.
	Context        context.Context // Set by the router after matching the route.
	ClientRequest  *http.Request   // Set by the router after matching the route.
	ClientResponse ResponseWriter  // Set by the router after matching the route.

	Upstream *Upstream             // Set by the upstream
	Endpoint loadbalancer.Endpoint // Set by the endpoint

	Error            error          // Set when aborting the context process anytime.
	UpstreamResponse *http.Response // Set by the upstream after forwarding the request.

	Data interface{} // The contex data that is set and used by the final user.

	next    Handler
	level   slog.LevelVar
	upreq   *http.Request
	queries url.Values
	cookies []*http.Cookie

	forwards    []func()
	respheaders []func()
	respbodys   []func()
}

// Reset resets the context to the initial state.
func (c *Context) Reset() {
	clear(c.forwards)
	clear(c.respbodys)
	clear(c.respheaders)

	*c = Context{
		forwards:    c.forwards[:0],
		respbodys:   c.respbodys[:0],
		respheaders: c.respheaders[:0],
	}
}

// ------------------------------------------------------------------------- //

func (c *Context) callbackOnForward()        { runcbs(c.forwards) }
func (c *Context) callbackOnResponseBody()   { runcbs(c.respbodys) }
func (c *Context) callbackOnResponseHeader() { runcbs(c.respheaders) }

func runcbs(cbs []func()) {
	for _, f := range cbs {
		f()
	}
}

// OnForward appends the callback function, which is called
// before upstream forwards the request.
func (c *Context) OnForward(cb func()) {
	c.forwards = append(c.forwards, cb)
}

// OnResponseHeader appends the callback function, which is called
// after setting the response header and before copying the response body
// from the upstream server to the client.
func (c *Context) OnResponseHeader(cb func()) {
	c.respheaders = append(c.respheaders, cb)
}

// OnResponseBody appends the callback function, which is called
// after copying the response body from the upstream server to the client.
func (c *Context) OnResponseBody(cb func()) {
	c.respbodys = append(c.respbodys, cb)
}

// UpstreamRequest returns the http request forwarded to the upstream server.
func (c *Context) UpstreamRequest() *http.Request {
	if c.upreq == nil {
		c.upreq = newUpstreamRequest(c)
	}
	return c.upreq
}

func (c *Context) upRequest() *http.Request { return c.upreq }

// IsAborted reports whether the context process is aborted.
func (c *Context) IsAborted() bool { return c.Error != nil }

// Abort sets the error informaion and aborts the context process.
func (c *Context) Abort(err error) { c.Error = err }

// ------------------------------------------------------------------------- //

// Cookie returns the cookie value by the name.
func (c *Context) Cookie(name string) string {
	for _, cookie := range c.Cookies() {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

// Cookies parses and returns the request header Cookie.
func (c *Context) Cookies() []*http.Cookie {
	if c.cookies == nil {
		c.cookies = c.ClientRequest.Cookies()
	}
	return c.cookies
}

// Queries parses and returns the query string.
func (c *Context) Queries() url.Values {
	if c.queries == nil {
		c.queries = c.ClientRequest.URL.Query()
	}
	return c.queries
}

func (c *Context) ClientIP() netip.Addr {
	if conn := nets.GetConnFromContext(c.Context); conn != nil {
		return assists.ConvertAddr(conn.RemoteAddr())
	}

	addr, _ := netip.ParseAddr(assists.TrimIP(c.ClientRequest.RemoteAddr))
	return addr
}

// RemoteAddr returns the remote address of the request.
func (c *Context) RemoteAddr() string { return c.ClientRequest.RemoteAddr }

// RequestID returns the request id of the request.
func (c *Context) RequestID() string { return c.ClientRequest.Header.Get(defaults.HeaderXRequestID) }

// NotFound reports whether the route is matched or not.
func (c *Context) NotFound() bool { return c.Route.matcher == nil }

// Enabled reports whether the level is enabled.
func (c *Context) Enabled(level slog.Level) bool {
	return level >= c.level.Level() && slog.Default().Enabled(c.Context, level)
}

// SetLogLevel sets the log level of the request context.
func (c *Context) SetLogLevel(level slog.Level) { c.level.Set(level) }

// SetResponseHandler resets the response handler.
func (c *Context) SetResponseHandler(h ResponseHandler) {
	c.Route.response = h
}

// SendResponse sends the response from the upstream server to the client.
func (c *Context) SendResponse() {
	if c.Route.response == nil {
		StdResponse(c, c.UpstreamResponse, c.Error)
	} else {
		c.Route.response(c, c.UpstreamResponse, c.Error)
	}
}

// SendRequest forwards the request to the upstream server.
func (c *Context) SendRequest() (*http.Response, error) {
	if c.Upstream.HttpClient != nil {
		return c.Upstream.HttpClient.Do(c.UpstreamRequest())
	}
	return DefaultHttpClient.Do(c.UpstreamRequest())
}
