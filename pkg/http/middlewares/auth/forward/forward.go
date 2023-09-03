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
	"errors"
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
		} else if config.Timeout < time.Second {
			config.Timeout *= time.Second
		}
		return ForwardAuth(name, auth.Priority, config)
	})
}

// Config is the configuration information to forward the authorization request
// to the authorization server.
type Config struct {
	// Required, either-or, and prefer route than url.
	Route string `json:"route" yaml:"route"`
	URL   string `json:"url" yaml:"url"`

	// Optional, Ignored if route is given
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
	auth := &forwardauth{conf: config}
	if auth.conf.URL == "" {
		return nil, fmt.Errorf("ForwardAuth: missing the url")
	}

	switch auth.conf.Method = strings.ToUpper(auth.conf.Method); auth.conf.Method {
	case http.MethodGet, http.MethodPost:
	case "":
		auth.conf.Method = http.MethodGet
	default:
		return nil, fmt.Errorf("ForwardAuth: unsupported the method '%s'", auth.conf.Method)
	}

	if auth.conf.Route == "" {
		var err error
		auth.req, err = http.NewRequest(auth.conf.Method, auth.conf.URL, nil)
		if err != nil {
			return nil, err
		}

		auth.authf = auth.authByReq
	} else {
		auth.authf = auth.authByRoute
	}

	if auth.conf.Timeout <= 0 {
		auth.conf.Timeout = time.Second * 3
	}

	auth.exactHeaders, auth.prefixHeaders = formatHeaders(auth.conf.Headers)
	auth.exactClientHeaders, auth.prefixClientHeaders = formatHeaders(auth.conf.ClientHeaders)
	auth.exactUpstreamHeaders, auth.prefixUpstreamHeaders = formatHeaders(auth.conf.UpstreamHeaders)

	return runtime.NewMiddleware(name, priority, auth.conf, func(next runtime.Handler) runtime.Handler {
		a := auth
		a.next = next
		return a.Handle
	}), nil
}

type forwardauth struct {
	next runtime.Handler
	conf Config
	req  *http.Request

	authf func(context.Context, *runtime.Context) (*http.Response, error)

	exactHeaders  map[string]struct{}
	prefixHeaders []string

	exactClientHeaders  map[string]struct{}
	prefixClientHeaders []string

	exactUpstreamHeaders  map[string]struct{}
	prefixUpstreamHeaders []string
}

func (a *forwardauth) Handle(c *runtime.Context) {
	ctx, cancel := context.WithTimeout(c.Context, a.conf.Timeout)
	defer cancel()

	resp, err := a.authf(ctx, c)
	if resp != nil { // resp is not eqial to nil when failing with 3xx.
		defer resp.Body.Close()
	}

	if err != nil {
		if a.conf.Degraded {
			a.next(c)
		} else {
			c.SendResponse(nil, runtime.ErrServiceUnavailable)
			c.Error = fmt.Errorf("forwardauth failed: url=%s, err=%w", a.conf.URL, err)
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
	slog.Error("forwardauth failed", "reqid", c.RequestID(),
		"method", a.conf.Method, "url", a.conf.URL, //"reqheaders", req.Header,
		"code", resp.StatusCode, "respheader", resp.Header,
		"err", buf.String())
	putbuffer(buf)

	if a.conf.Degraded {
		a.next(c)
	} else {
		copyHeaders(c.ClientResponse.Header(), resp.Header, a.exactClientHeaders, a.prefixClientHeaders)
		c.SendResponse(nil, runtime.ErrUnauthorized.WithError(errAuthFailure))
		c.Error = fmt.Errorf("forwardauth failed: method=%s, url=%s, code=%d",
			a.conf.Method, a.conf.URL, resp.StatusCode)
	}
}

func (a *forwardauth) authByReq(ctx context.Context, c *runtime.Context) (resp *http.Response, err error) {
	req := a.req.Clone(ctx)
	a.updateAuthReq(c, req)

	switch {
	case a.conf.Client != nil:
		resp, err = a.conf.Client.Do(req)

	case runtime.DefaultHttpClient != nil:
		resp, err = runtime.DefaultHttpClient.Do(req)

	default:
		resp, err = http.DefaultClient.Do(req)
	}

	if err != nil {
		slog.Error("forwardauth failed", "reqid", c.RequestID(),
			"method", a.conf.Method, "url", a.conf.URL, "err", err)
	}

	return
}

func (a *forwardauth) authByRoute(ctx context.Context, c *runtime.Context) (*http.Response, error) {
	// route, ok := runtime.DefaultRouter.GetRoute(a.conf.Route)
	// if !ok {
	// 	return nil, fmt.Errorf("not found the route '%s'", a.conf.Route)
	// }

	// TODO:
	return nil, nil
}

func (a *forwardauth) updateAuthReq(c *runtime.Context, req *http.Request) {
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
}

var bufpool = &sync.Pool{New: func() any { return bytes.NewBuffer(make([]byte, 0, 128)) }}

func getbuffer() *bytes.Buffer  { return bufpool.Get().(*bytes.Buffer) }
func putbuffer(b *bytes.Buffer) { b.Reset(); bufpool.Put(b) }

var errAuthFailure = errors.New("auth failed")

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
