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
	"context"
	"net"
	"net/http"
	"time"

	"github.com/xgfone/go-apigateway/http/core"
)

// Pre-define some upstream hosts.
const (
	HostClient = "$client"
	HostServer = "$server"
)

// DefaultHttpClient is the default global http client
// to forward the request to the upstream server.
var DefaultHttpClient = newClient()

// EnableRedirectHttpClient is the same as DefaultHttpClient,
// which enables redirecting and returns the 3xx response.
var EnableRedirectHttpClient = newClient()

// DisableRedirectHttpClient is the same as DefaultHttpClient,
// which disables redirecting and returns the 3xx response.
var DisableRedirectHttpClient = newDisableRedirectClient()

// Send sends the http request with context and returns the http response.
func Send(c *core.Context, r *http.Request) (*http.Response, error) {
	if c.Client != nil {
		return c.Client.Do(r)
	}
	return DefaultHttpClient.Do(r)
}

func disableRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

func newDisableRedirectClient() *http.Client {
	client := *DefaultHttpClient
	client.CheckRedirect = disableRedirect
	return &client
}

func newClient() *http.Client {
	return &http.Client{Transport: newTransport()}
}

func newTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport)
	t.DialContext = dialContext
	t.MaxIdleConns = 0
	t.MaxIdleConnsPerHost = 100
	t.TLSHandshakeTimeout = time.Second * 2
	// t.IdleConnTimeout = time.Second * 90
	return t
}

func dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d := net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}
	return d.DialContext(ctx, network, addr)
}

func cloneRequest(orig *http.Request) *http.Request {
	req := orig.Clone(orig.Context())
	req.RequestURI = "" // Pretend to be a client request.
	//req.URL.Host = "" // Dial to the backend http endpoint.
	return req
}

// NewRequest returns a new upstream request by the context.
func newRequest(c *core.Context) (req *http.Request) {
	if c.ClientRequest == nil {
		panic("NewUpstreamRequest: ClientRequest is nil")
	}

	if c.ClientResponse == nil { // call directly
		req = c.ClientRequest
	} else { // server forward
		req = cloneRequest(c.ClientRequest)
		req.URL.User = nil           // Clear the basic auth.
		req.Close = false            // Enable the keepalive
		req.Header.Del("Connection") // Enable the keepalive
	}

	return
}
