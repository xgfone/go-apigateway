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

// Package matcher provides a route matcher and building.
package matcher

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/router"
	"github.com/xgfone/go-defaults"
	"github.com/xgfone/go-defaults/assists"
)

var (
	// GetHost is used to customize the host, which must be lower case.
	GetHost = func(r *http.Request) string {
		if r.TLS != nil && r.TLS.ServerName != "" {
			return strings.ToLower(r.TLS.ServerName)
		}
		return strings.ToLower(r.Host)
	}

	// GetPath is used to customize the path.
	GetPath = func(r *http.Request) string {
		return strings.TrimRight(r.URL.Path, "/")
	}

	// GetClientIP is used to customize the client ip.
	GetClientIP = func(r *http.Request) netip.Addr {
		return defaults.GetClientIP(r.Context(), r)
	}

	// GetServerIP is used to customize the server ip.
	GetServerIP = func(r *http.Request) netip.Addr {
		addr := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
		return assists.ConvertAddr(addr)
	}
)

type Matcher interface {
	router.Matcher
	fmt.Stringer
	Priority() int
}

// Matcher is used to match a http request.
type matcher struct {
	match router.MatcherFunc
	desc  string
	prio  int
}

func (m matcher) String() string             { return m.desc }
func (m matcher) Priority() int              { return m.prio }
func (m matcher) Match(c *core.Context) bool { return m.match(c) }

// New returns a request matcher.
func New(prio int, desc string, match router.MatcherFunc) Matcher {
	return matcher{prio: prio, desc: desc, match: match}
}
