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

// Package core provides some core runtime functions.
package core

import (
	"context"
	"net/http"
	"net/netip"
	"net/url"
	"sync"
	"time"

	"github.com/xgfone/go-apigateway/nets"
	"github.com/xgfone/go-defaults"
	"github.com/xgfone/go-defaults/assists"
	"github.com/xgfone/go-loadbalancer"
)

var DefaultCapSize = 4

var ctxpool = &sync.Pool{New: func() any { return NewContext() }}

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

// Context represents a runtime context.
type Context struct {
	Context context.Context
	Next    Handler

	RouteId    string // Set by the router after matching the route.
	UpstreamId string // Set by the router after matching the route.

	// For Client
	Client         *http.Client
	ClientRequest  *http.Request  // Set by the router when serving the request.
	ClientResponse ResponseWriter // Set by the router when serving the request.
	Responser      Responser      // Set by the router after matching the route.

	// For Upstream
	Upstream         interface{}
	UpstreamRequest  *http.Request
	UpstreamResponse *http.Response // Set by the upstream after forwarding the request.
	ForwardTimeout   time.Duration
	Endpoint         loadbalancer.Endpoint // Set by the endpoint

	IsAborted bool
	Error     error          // Set when aborting the context process anytime.
	Data      interface{}    // The contex data that is set and used by the final user.
	Kvs       map[string]any // The interim context key-value cache.

	// Cache to avoid to parse them twice.
	queries url.Values
	cookies []*http.Cookie

	// Callbacks
	forwards    []func()
	respheaders []func()
}

func NewContext() *Context {
	return &Context{
		Kvs:         make(map[string]any, DefaultCapSize),
		forwards:    make([]func(), 0, DefaultCapSize),
		respheaders: make([]func(), 0, DefaultCapSize),
	}
}

// ------------------------------------------------------------------------- //

// Reset resets the context to the initial state.
func (c *Context) Reset() {
	clear(c.Kvs)
	clear(c.forwards)
	clear(c.respheaders)
	*c = Context{Kvs: c.Kvs, forwards: c.forwards[:0], respheaders: c.respheaders[:0]}
}

// Abort sets the error informaion and aborts the context process.
func (c *Context) Abort(err error) {
	c.IsAborted = true
	c.Error = err
}

// ------------------------------------------------------------------------- //

func runcbs(cbs []func()) {
	for _, f := range cbs {
		f()
	}
}

// CallbackOnForward calls the callback functions added by OnForward.
func (c *Context) CallbackOnForward() { runcbs(c.forwards) }

// CallbackOnResponseHeader calls the callback functions added by OnResponseHeader.
func (c *Context) CallbackOnResponseHeader() { runcbs(c.respheaders) }

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

// ------------------------------------------------------------------------- //

// SendResponse sends the response from the upstream server to the client.
func (c *Context) SendResponse() {
	if c.Responser == nil {
		StdResponse(c, c.UpstreamResponse, c.Error)
	} else {
		c.Responser.Respond(c, c.UpstreamResponse, c.Error)
	}
}

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
	if conn := nets.GetConnFromContext(c.ClientRequest.Context()); conn != nil {
		return assists.ConvertAddr(conn.RemoteAddr())
	}

	addr, _ := netip.ParseAddr(assists.TrimPort(c.ClientRequest.RemoteAddr))
	return addr
}

// RemoteAddr returns the remote address of the request.
func (c *Context) RemoteAddr() string { return c.ClientRequest.RemoteAddr }

// RequestID returns the request id of the request.
func (c *Context) RequestID() string { return c.ClientRequest.Header.Get(defaults.HeaderXRequestID) }
