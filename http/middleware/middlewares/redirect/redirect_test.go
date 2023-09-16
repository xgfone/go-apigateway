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

package redirect

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
)

func TestRedirect(t *testing.T) {
	m, err := Redirect(Config{HttpToHttps: true})
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/path?key=value", nil)
	req.RequestURI = strings.TrimPrefix(req.RequestURI, "http://127.0.0.1")

	c := core.AcquireContext(req.Context())
	c.ClientResponse = core.AcquireResponseWriter(rec)
	c.ClientRequest = req

	/// http -> https
	m.Handler(func(c *core.Context) { panic("redirect failed") })(c)
	if rec.Code != 302 {
		t.Errorf("expect status code 302, but got %d", rec.Code)
	}
	expectloc := "https://127.0.0.1/path?key=value"
	if loc := rec.Header().Get("Location"); loc != expectloc {
		t.Errorf("expect Location '%s', but got '%s'", expectloc, loc)
	}

	/// https -> handler
	rec = httptest.NewRecorder()
	c.ClientResponse = core.AcquireResponseWriter(rec)
	c.ClientRequest.TLS = new(tls.ConnectionState)
	m.Handler(func(c *core.Context) { c.ClientResponse.WriteHeader(204) })(c)
	if rec.Code != 204 {
		t.Errorf("expect status code 204, but got %d", rec.Code)
	}

	/// Location without Query
	m, err = Redirect(Config{Location: "/redirect"})
	if err != nil {
		t.Fatal(err)
	}

	rec = httptest.NewRecorder()
	c.ClientResponse = core.AcquireResponseWriter(rec)
	m.Handler(func(c *core.Context) { panic("redirect failed") })(c)
	if rec.Code != 302 {
		t.Errorf("expect status code 302, but got %d", rec.Code)
	}
	expectloc = "/redirect"
	if loc := rec.Header().Get("Location"); loc != expectloc {
		t.Errorf("expect Location '%s', but got '%s'", expectloc, loc)
	}

	/// Location with Query
	m, err = Redirect(Config{Location: "/redirect", AppendQuery: true})
	if err != nil {
		t.Fatal(err)
	}

	rec = httptest.NewRecorder()
	c.ClientResponse = core.AcquireResponseWriter(rec)
	m.Handler(func(c *core.Context) { panic("redirect failed") })(c)
	if rec.Code != 302 {
		t.Errorf("expect status code 302, but got %d", rec.Code)
	}
	expectloc = "/redirect?key=value"
	if loc := rec.Header().Get("Location"); loc != expectloc {
		t.Errorf("expect Location '%s', but got '%s'", expectloc, loc)
	}
}
