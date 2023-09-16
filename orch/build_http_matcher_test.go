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

package orch

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
)

func TestBuildClientIP(t *testing.T) {
	expect := "ClientIps(`127.0.0.0/8`)"
	config := HttpMatcher{ClientIps: []string{"127.0.0.0/8"}}
	if m, err := config.buildClientIP(); err != nil {
		t.Error(err)
	} else if desc := m.String(); desc != expect {
		t.Errorf("expect '%s', but got '%s'", expect, desc)
	}

	expect = "ClientIps(`127.0.0.0/8`,`192.168.0.0/24`)"
	config.ClientIps = append(config.ClientIps, "192.168.0.0/24")
	if m, err := config.buildClientIP(); err != nil {
		t.Error(err)
	} else if desc := m.String(); desc != expect {
		t.Errorf("expect '%s', but got '%s'", expect, desc)
	}
}

func TestBuildMatcher(t *testing.T) {
	m, err := HttpMatcher{
		Methods: []string{"GET"},

		Paths:        []string{"/path///"},
		PathPrefixes: []string{"/p///"},

		Hosts: []string{"*.example.com"},

		Headers: map[string]string{"X-Test": "1"},
		Queries: map[string]string{"key": "value"},

		ClientIps: []string{"127.0.0.0/8"},
		ServerIps: []string{"127.0.0.0/8"},
	}.Build()
	if err != nil {
		t.Fatal(err)
	}

	_ = m.Priority()

	req := &http.Request{
		Method:     "GET",
		URL:        &url.URL{Path: "/path", RawQuery: "key=value"},
		Host:       "www.example.com",
		RemoteAddr: "127.0.0.1",
		Header:     http.Header{"X-Test": []string{"1"}},
	}
	localaddr := &net.TCPAddr{Port: 80, IP: net.ParseIP("127.0.0.1")}
	ctx := context.WithValue(req.Context(), http.LocalAddrContextKey, localaddr)
	req = req.WithContext(ctx)
	c := core.AcquireContext(ctx)

	c.Reset()
	c.Context = ctx
	c.ClientRequest = req
	testmatcherbuilder(t, "mbsuccess", m.Match(c), true)

	c.Reset()
	c.Context = ctx
	c.ClientRequest = req.Clone(ctx)
	c.ClientRequest.Method = "PUT"
	testmatcherbuilder(t, "mbfail_method", m.Match(c), false)

	c.Reset()
	c.Context = ctx
	c.ClientRequest = req.Clone(ctx)
	c.ClientRequest.URL.Path = "/p"
	testmatcherbuilder(t, "mbfail_path", m.Match(c), false)

	c.Reset()
	c.Context = ctx
	c.ClientRequest = req.Clone(ctx)
	c.ClientRequest.URL.Path = "/"
	testmatcherbuilder(t, "mbfail_pathprefix", m.Match(c), false)

	c.Reset()
	c.Context = ctx
	c.ClientRequest = req.Clone(ctx)
	c.ClientRequest.Host = "localhost"
	testmatcherbuilder(t, "mbfail_host", m.Match(c), false)

	c.Reset()
	c.Context = ctx
	c.ClientRequest = req.Clone(ctx)
	c.ClientRequest.Header.Set("X-Test", "0")
	testmatcherbuilder(t, "mbfail_header", m.Match(c), false)

	c.Reset()
	c.Context = ctx
	c.ClientRequest = req.Clone(ctx)
	c.ClientRequest.URL.RawQuery = ""
	testmatcherbuilder(t, "mbfail_query", m.Match(c), false)

	c.Reset()
	c.Context = ctx
	c.ClientRequest = req.Clone(ctx)
	c.ClientRequest.RemoteAddr = "192.168.1.1"
	testmatcherbuilder(t, "mbfail_clientip", m.Match(c), false)

	c.Reset()
	c.Context = ctx
	c.ClientRequest = req.Clone(ctx)
	localaddr.IP = net.ParseIP("192.168.1.1")
	testmatcherbuilder(t, "mbfail_serverip", m.Match(c), false)
}

func testmatcherbuilder(t *testing.T, prefix string, result, expect bool) {
	if expect != result {
		t.Errorf("%s: expect %v, but got %v", prefix, expect, result)
	}
}
