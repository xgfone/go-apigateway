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

// Package forward provides a middleware that forward the auth to the external server.
package forward

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/xgfone/go-apigateway/pkg/http/middlewares/auth"
	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-binder"
)

func init() {
	runtime.RegisterMiddlewareBuilder("forwardauth", func(name string, conf map[string]any) (runtime.Middleware, error) {
		var config Config
		if err := binder.BindStructToMap(&config, "json", conf); err != nil {
			return nil, err
		}
		return ForwardAuth(name, auth.Priority, config)
	})
}

// Config is the configuration information to forward the authorization request
// to the authorization server.
type Config struct {
	// Required
	URL string `json:"url,omitempty" yaml:"url,omitempty"`

	// Optional, if exists, use the upstream as the transport.
	Upstream string `json:"upstream,omitempty" yaml:"upstream,omitempty"`

	// Optional
	Method  string   `json:"method,omitempty" yaml:"method,omitempty"`   // Default: GET, one of GET or POST
	Headers []string `json:"headers,omitempty" yaml:"headers,omitempty"` // Default: nil
	/// Extra Request Headers:
	// X-Forwarded-Proto:   Scheme
	// X-Forwarded-Method:  HTTP Method
	// X-Forwarded-Host:    Host
	// X-Forwarded-Uri:     URI
	// X-Forwarded-For:     RemoteIP

	// Timeout to get the auth result from the authorization server.
	//
	// Default: 3s
	Timeout time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// These is the response headers returned by the authorization server.
	// UpstreamHeaders are forwarded to the upstream backend server if auth succeeded.
	// ClientHeaders are forwarded to the client if auth failed.
	//
	// Support that the last character is "*" as the prefix matching,
	// Such as "X-User-*".
	//
	// Default: nil
	UpstreamHeaders []string `json:"upstreamHeaders,omitempty" yaml:"upstreamHeaders,omitempty"`
	ClientHeaders   []string `json:"clientHeaders,omitempty" yaml:"clientHeaders,omitempty"`

	// If true and auth failure, still forward the request to upstream backend server.
	//
	// Default: false
	Degraded bool `json:"degraded,omitempty" yaml:"degraded,omitempty"`

	// Default: http.DefaultClient
	Client *http.Client `json:"-" yaml:"-"`
}

// ForwardAuth returns a new middleware that forwards the authentication
// to the external server.
func ForwardAuth(name string, priority int, config Config) (runtime.Middleware, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("ForwardAuth: missing the url")
	}

	switch config.Method = strings.ToUpper(config.Method); config.Method {
	case http.MethodGet, http.MethodPost:
	case "":
		config.Method = http.MethodGet
	default:
		return nil, fmt.Errorf("ForwardAuth: unsupported the method '%s'", config.Method)
	}

	req, err := http.NewRequest(config.Method, config.URL, nil)
	if err != nil {
		return nil, err
	}

	if config.Timeout <= 0 {
		config.Timeout = time.Second * 3
	}

	auth := forwardauth{
		name: name,
		upid: config.Upstream,
		req:  req,

		url:      config.URL,
		degraded: config.Degraded,
		timeout:  config.Timeout,
		client:   config.Client,
	}

	auth.exactHeaders, auth.prefixHeaders = formatHeaders(config.Headers)
	auth.exactClientHeaders, auth.prefixClientHeaders = formatHeaders(config.ClientHeaders)
	auth.exactUpstreamHeaders, auth.prefixUpstreamHeaders = formatHeaders(config.UpstreamHeaders)

	return runtime.NewMiddleware(name, priority, config, func(next runtime.Handler) runtime.Handler {
		auth := auth.with(next)
		return auth.Handle
	}), nil
}

type forwardauth struct {
	next runtime.Handler

	url      string
	name     string
	upid     string
	degraded bool
	timeout  time.Duration
	client   *http.Client
	req      *http.Request

	exactHeaders  map[string]struct{}
	prefixHeaders []string

	exactClientHeaders  map[string]struct{}
	prefixClientHeaders []string

	exactUpstreamHeaders  map[string]struct{}
	prefixUpstreamHeaders []string
}

func (a forwardauth) with(next runtime.Handler) forwardauth {
	a.next = next
	return a
}

func (a *forwardauth) Handle(c *runtime.Context) {
	ctx, cancel := context.WithTimeout(c.Context, a.timeout)
	defer cancel()

	req := a.req.Clone(ctx)

	// 1. Copy the headers from the original request headers.
	copyHeaders(req.Header, c.ClientRequest.Header, a.exactHeaders, a.prefixHeaders)

	// 2. Add the extra headers.
	if c.ClientRequest.TLS == nil {
		req.Header.Set("X-Forwarded-Proto", "http")
	} else {
		req.Header.Set("X-Forwarded-Proto", "https")
	}
	req.Header.Set("X-Forwarded-Method", c.ClientRequest.Method)
	req.Header.Set("X-Forwarded-Host", c.ClientRequest.Host)
	req.Header.Set("X-Forwarded-Uri", c.ClientRequest.RequestURI)
	req.Header.Set("X-Forwarded-For", c.ClientRequest.RemoteAddr)

	// 3. Send the request to authorization server
	var err error
	var resp *http.Response
	switch {
	case a.upid != "":
		up := runtime.GetUpstream(a.upid)
		if up == nil {
			err = fmt.Errorf("not found the upstream '%s'", a.upid)
			break
		}

		newc := runtime.AcquireContext(c.Context)
		newc.ClientRequest = req
		up.Forward(newc)
		resp, err = newc.UpstreamResponse, newc.Error
		runtime.ReleaseContext(newc)

	case a.client != nil:
		resp, err = a.client.Do(req)

	case runtime.DefaultHttpClient != nil:
		resp, err = runtime.DefaultHttpClient.Do(req)

	default:
		resp, err = http.DefaultClient.Do(req)
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		slog.Error("fail to forward auth", "reqid", c.RequestID(),
			"method", req.Method, "url", a.url, "upstream", a.upid, "err", err)

		if a.degraded {
			a.next(c)
		} else {
			err := fmt.Errorf("fail to forward auth: url=%s, err=%w", a.url, err)
			c.Abort(err)
		}

		return
	}

	// Success
	if resp.StatusCode < 300 {
		// Add the headers into the request when forwarding it.
		c.OnForward(func() {
			copyHeaders(c.UpstreamRequest().Header, resp.Header, a.exactUpstreamHeaders, a.prefixUpstreamHeaders)
		})
		a.next(c)
		return
	}

	// Failure

	buf := getbuffer()
	_, _ = io.Copy(buf, resp.Body)
	slog.Error("fail to forward auth",
		"reqid", c.RequestID(), "route", c.Route.Route.Id,
		"authmethod", req.Method, "authurl", a.url, "authreqheaders", req.Header,
		"authrespcode", resp.StatusCode, "authrespheader", resp.Header,
		"authrespbody", buf.String())
	putbuffer(buf)

	if a.degraded {
		a.next(c)
	} else {
		copyHeaders(c.ClientResponse.Header(), resp.Header, a.exactClientHeaders, a.prefixClientHeaders)
		c.Abort(runtime.ErrUnauthorized)
	}
}

var bufpool = &sync.Pool{New: func() any { return bytes.NewBuffer(make([]byte, 0, 128)) }}

func getbuffer() *bytes.Buffer  { return bufpool.Get().(*bytes.Buffer) }
func putbuffer(b *bytes.Buffer) { b.Reset(); bufpool.Put(b) }

func formatHeaders(headers []string) (exactHeaders map[string]struct{}, prefixHeaders []string) {
	var elen, plen int
	for _, key := range headers {
		if strings.HasSuffix(key, "*") {
			plen++
		} else {
			elen++
		}
	}

	if elen == 0 {
		prefixHeaders = make([]string, 0, plen)
		for _, key := range headers {
			prefixHeaders = append(prefixHeaders, http.CanonicalHeaderKey(key[:len(key)-1]))
		}
		return
	}

	exactHeaders = make(map[string]struct{}, elen)
	prefixHeaders = make([]string, 0, plen)

	for _, key := range headers {
		if strings.HasSuffix(key, "*") {
			prefixHeaders = append(prefixHeaders, http.CanonicalHeaderKey(key[:len(key)-1]))
		} else {
			exactHeaders[http.CanonicalHeaderKey(key)] = struct{}{}
		}
	}

	return
}

func copyHeaders(dst, src http.Header, exactKeys map[string]struct{}, prefixKeys []string) {
	for key, values := range src {
		if _, ok := exactKeys[key]; ok {
			dst[key] = values
			continue
		}

		ok := slices.ContainsFunc(prefixKeys, func(s string) bool {
			return strings.HasPrefix(key, s)
		})
		if ok {
			dst[key] = values
		}
	}
}
