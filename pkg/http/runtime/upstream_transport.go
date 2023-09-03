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
	"net"
	"net/http"
	"time"
)

var (
	// NewRequest is used to return a new from the original request.
	NewRequest = newRequest

	// DefaultHttpClient is the default global http client
	// to forward the request to the upstream server.
	DefaultHttpClient = newClient()
)

func newClient() *http.Client {
	return &http.Client{Transport: newTransport()}
}

func newTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport)
	t.DialContext = dialContext
	t.MaxIdleConns = 1000
	t.MaxIdleConnsPerHost = 100
	t.TLSHandshakeTimeout = time.Second * 3
	// t.IdleConnTimeout = time.Second * 90
	return t
}

func dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d := net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}
	return d.DialContext(ctx, network, addr)
}

func newRequest(orig *http.Request) *http.Request {
	req := orig.Clone(orig.Context())
	req.RequestURI = "" // Pretend to be a client request.
	//req.URL.Host = "" // Dial to the backend http endpoint.
	return req
}

func newUpstreamRequest(c *Context) *http.Request {
	req := NewRequest(c.ClientRequest)
	req.URL.User = nil           // Clear the basic auth.
	req.Close = false            // Enable the keepalive
	req.Header.Del("Connection") // Enable the keepalive

	// At upstream or endpoint forwarding point.
	if c.Upstream != nil {
		updateUpstreamRequest(c, req)
	}
	return req
}
