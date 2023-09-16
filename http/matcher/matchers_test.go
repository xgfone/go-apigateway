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

package matcher

import (
	"context"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
)

func TestMatchers(t *testing.T) {
	ms := []Matcher{
		New(1, "Method(`GET`)", func(c *core.Context) bool { return c.ClientRequest.Method == "GET" }),
		New(1, "Path(`/`)", func(c *core.Context) bool { return c.ClientRequest.URL.Path == "/" }),
		New(3, "ClientIp(`127.0.0.1`)", func(c *core.Context) bool { return c.ClientRequest.RemoteAddr == "127.0.0.1" }),
		New(2, "Host(`localhost`)", func(c *core.Context) bool { return c.ClientRequest.Host == "localhost" }),
	}

	if p := And(ms...).Priority(); p != 7 {
		t.Errorf("expect priority %d, but got %d", 7, p)
	}

	Sort(ms)
	descs := make([]string, len(ms))
	for i, m := range ms {
		descs[i] = m.String()
	}
	expects := []string{
		"ClientIp(`127.0.0.1`)",
		"Host(`localhost`)",
		"Method(`GET`)",
		"Path(`/`)",
	}
	if !reflect.DeepEqual(expects, descs) {
		t.Errorf("expect %v, but got %v", expects, descs)
	}

	c := core.AcquireContext(context.Background())
	c.ClientRequest = &http.Request{
		Method:     "GET",
		URL:        &url.URL{Path: "/"},
		Host:       "localhost",
		RemoteAddr: "127.0.0.1",
	}
	if !And(ms...).Match(c) {
		t.Errorf("expect matched, but got false")
	}

	c.ClientRequest.RemoteAddr = "127.0.0.1:80"
	if And(ms...).Match(c) {
		t.Error("expect not matched, but got true")
	}

	andexpect := "(ClientIp(`127.0.0.1`) && Host(`localhost`) && Method(`GET`) && Path(`/`))"
	if desc := And(ms...).String(); desc != andexpect {
		t.Errorf("expect '%s', but got '%s'", andexpect, desc)
	}

	orexpect := "(ClientIp(`127.0.0.1`) || Host(`localhost`) || Method(`GET`) || Path(`/`))"
	if desc := Or(ms...).String(); desc != orexpect {
		t.Errorf("expect '%s', but got '%s'", orexpect, desc)
	}
}

func TestRemoveParentheses(t *testing.T) {
	if s := RemoveParentheses("abc"); s != "abc" {
		t.Errorf("expect '%s', but got '%s'", "abc", s)
	}
	if s := RemoveParentheses("(abc"); s != "(abc" {
		t.Errorf("expect '%s', but got '%s'", "(abc", s)
	}
	if s := RemoveParentheses("abc)"); s != "abc)" {
		t.Errorf("expect '%s', but got '%s'", "abc)", s)
	}
	if s := RemoveParentheses("(abc)"); s != "abc" {
		t.Errorf("expect '%s', but got '%s'", "abc", s)
	}
}
