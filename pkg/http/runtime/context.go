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
	"net"
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
		upcbs: make([]func(), 0, 2),
		rhcbs: make([]func(), 0, 4),
		rbcbs: make([]func(), 0, 2),
	}
}}

// AcquireContext acquires a context from the pool.
func AcquireContext() *Context { return ctxpool.Get().(*Context) }

// ReleaseContext releases the context to the pool.
func ReleaseContext(c *Context) {
	clear(c.rhcbs)
	clear(c.rbcbs)

	*c = Context{rhcbs: c.rhcbs[:0], rbcbs: c.rbcbs[:0]}
	ctxpool.Put(c)
}

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
	Route Route // Set by the router after matching the route.

	Context        context.Context // Set by the router
	ClientRequest  *http.Request   // Set by the router
	ClientResponse ResponseWriter  // Set by the router

	Upstream *Upstream             // Set by the upstream
	Endpoint loadbalancer.Endpoint // Set by the endpoint

	Error            error          // Set by the upstream after forwarding the request.
	UpstreamResponse *http.Response // Set by the upstream after forwarding the request.

	Data interface{} // The contex data that is set and used by the final user.

	next    Handler
	level   slog.LevelVar
	upreq   *http.Request
	queries url.Values
	cookies []*http.Cookie
	upcbs   []func()
	rhcbs   []func()
	rbcbs   []func()
}

func (c *Context) upRequest() *http.Request  { return c.upreq }
func (c *Context) callbackOnForward()        { runcbs(c.upcbs) }
func (c *Context) callbackOnResponseBody()   { runcbs(c.rbcbs) }
func (c *Context) callbackOnResponseHeader() { runcbs(c.rhcbs) }

func runcbs(cbs []func()) {
	for _, f := range cbs {
		f()
	}
}

// OnForward appends the callback function, which is called
// before upstream forwards the request.
func (c *Context) OnForward(cb func()) {
	c.upcbs = append(c.upcbs, cb)
}

// OnResponseHeader appends the callback function, which is called
// after setting the response header and before copying the response body
// from the upstream server to the client.
func (c *Context) OnResponseHeader(cb func()) {
	c.rhcbs = append(c.rhcbs, cb)
}

// OnResponseBody appends the callback function, which is called
// after copying the response body from the upstream server to the client.
func (c *Context) OnResponseBody(cb func()) {
	c.rbcbs = append(c.rbcbs, cb)
}

// UpstreamRequest returns the http request forwarded to the upstream server.
func (c *Context) UpstreamRequest() *http.Request {
	if c.upreq == nil {
		c.upreq = newUpstreamRequest(c)
	}
	return c.upreq
}

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
		ip := conn.RemoteAddr().(*net.TCPAddr).IP
		if ipv4 := ip.To4(); ipv4 != nil {
			ip = ipv4
		}
		addr, _ := netip.AddrFromSlice(ip)
		return addr
		// return assists.ConvertAddr(conn.RemoteAddr())
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
func (c *Context) SendResponse(resp *http.Response, err error) {
	c.UpstreamResponse, c.Error = resp, err
	if c.Route.response == nil {
		StdResponse(c, resp, err)
	} else {
		c.Route.response(c, resp, err)
	}
}

// SendRequest forwards the request to the upstream server.
func (c *Context) SendRequest() (*http.Response, error) {
	if c.Upstream.HttpClient != nil {
		return c.Upstream.HttpClient.Do(c.UpstreamRequest())
	}
	return DefaultHttpClient.Do(c.UpstreamRequest())
}
