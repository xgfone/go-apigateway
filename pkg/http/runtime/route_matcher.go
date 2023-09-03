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
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/textproto"
	"slices"
	"strings"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
	"github.com/xgfone/go-apigateway/pkg/nets"
	"github.com/xgfone/go-defaults"
	"github.com/xgfone/go-defaults/assists"
)

var (
	// GetHost is used to customize the host, which must be lower case.
	GetHost = func(r *http.Request) string { return strings.ToLower(r.Host) }

	// GetPath is used to customize the path.
	GetPath = func(r *http.Request) string { return r.URL.Path }

	// GetClientIP is used to customize the client ip.
	GetClientIP = func(r *http.Request) netip.Addr {
		return defaults.GetClientIP(r.Context(), r)
	}

	// GetServerIP is used to customize the server ip.
	GetServerIP = func(r *http.Request) netip.Addr {
		addr := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
		return netaddr2netipaddr(addr)
	}
)

// Copy from assists.ConvertAddr
func netaddr2netipaddr(addr net.Addr) netip.Addr {
	switch v := addr.(type) {
	case *net.TCPAddr:
		addr, _ := netip.AddrFromSlice(v.IP)
		return addr

	case *net.UDPAddr:
		addr, _ := netip.AddrFromSlice(v.IP)
		return addr

	default:
		addr, _ := netip.ParseAddr(assists.TrimIP(v.String()))
		return addr
	}
}

// Matcher is used to check whether the route rule matches the request.
type matcher interface {
	// Priority is the priority of the matcher.
	//
	// The bigger the value, the higher the priority.
	Priority() int

	// Match is used to check whether the rule matches the request.
	Match(*Context) bool
}

type matchers []matcher

func (ms matchers) Priority() (p int) {
	for _, m := range ms {
		p += m.Priority()
	}
	return
}

func (ms matchers) Match(c *Context) bool {
	for _, m := range ms {
		if !m.Match(c) {
			return false
		}
	}
	return true
}

type _matcher struct {
	priority  int
	matchfunc func(*Context) bool
}

func newMatcher(p int, f func(c *Context) bool) matcher { return _matcher{p, f} }
func (m _matcher) Match(c *Context) bool                { return m.matchfunc(c) }
func (m _matcher) Priority() int                        { return m.priority }

// ----------------------------------------------------------------------- //

func buildRouteMatcher(r dynamicconfig.Route) (matcher, error) {
	ms := make(matchers, 0, 4)

	ms = _appendmatcher(ms, buildRouteMatcherForMethod(r.Matcher))
	ms = _appendmatcher(ms, buildRouteMatcherForPath(r.Matcher))
	ms = _appendmatcher(ms, buildRouteMatcherForPathPrefix(r.Matcher))
	ms = _appendmatcher(ms, buildRouteMatcherForHost(r.Matcher))
	ms = _appendmatcher(ms, buildRouteMatcherForHeaders(r.Matcher))
	ms = _appendmatcher(ms, buildRouteMatcherForQueries(r.Matcher))

	clientIPMatcher, err := buildRouteMatcherForClientIP(r.Matcher)
	if err != nil {
		return nil, err
	}
	ms = _appendmatcher(ms, clientIPMatcher)

	serverIPMatcher, err := buildRouteMatcherForServerIP(r.Matcher)
	if err != nil {
		return nil, err
	}
	ms = _appendmatcher(ms, serverIPMatcher)

	if len(ms) == 0 {
		return nil, errors.New("no route matcher")
	}
	return ms, nil
}

func _appendmatcher(ms matchers, m matcher) matchers {
	if m != nil {
		ms = append(ms, m)
	}
	return ms
}

const (
	priorityHeader     = 2
	priorityQuery      = 2
	priorityClientIP   = 3
	priorityServerIP   = 3
	priorityMethod     = 4
	priorityPathPrefix = 5
	priorityPath       = 6
	priorityHost       = 8
)

func mergestrings(s string, ss []string) []string {
	_len := len(ss)
	if _len == 0 {
		if s == "" {
			return nil
		}
		return []string{s}
	} else if s == "" {
		return ss
	}

	vs := make([]string, _len+1)
	copy(vs, ss)
	vs[_len] = s
	return vs
}

func buildRouteMatcherForMethod(m dynamicconfig.Matcher) matcher {
	switch methods := mergestrings(m.Method, m.Methods); len(methods) {
	case 0:
		return nil

	case 1:
		method := strings.ToUpper(methods[0])
		return newMatcher(priorityMethod, func(c *Context) bool {
			return c.ClientRequest.Method == method
		})

	default:
		for i, m := range methods {
			methods[i] = strings.ToUpper(m)
		}
		return newMatcher(priorityMethod, func(c *Context) bool {
			return slices.Contains(methods, c.ClientRequest.Method)
		})

	}
}

func buildRouteMatcherForPath(m dynamicconfig.Matcher) matcher {
	switch paths := mergestrings(m.Path, m.Paths); len(paths) {
	case 0:
		return nil

	case 1:
		path := paths[0]
		return newMatcher(priorityPath, func(c *Context) bool {
			return GetPath(c.ClientRequest) == path
		})

	default:
		return newMatcher(priorityPath, func(c *Context) bool {
			return slices.Contains(paths, GetPath(c.ClientRequest))
		})

	}
}

func buildRouteMatcherForPathPrefix(m dynamicconfig.Matcher) matcher {
	switch prefixs := mergestrings(m.PathPrefix, m.PathPrefixs); len(prefixs) {
	case 0:
		return nil

	case 1:
		prefix := prefixs[0]
		return newMatcher(priorityPathPrefix, func(c *Context) bool {
			return strings.HasPrefix(GetPath(c.ClientRequest), prefix)
		})

	default:
		return newMatcher(priorityPathPrefix, func(c *Context) bool {
			path := GetPath(c.ClientRequest)
			return slices.IndexFunc(prefixs, func(s string) bool {
				return strings.HasPrefix(path, s)
			}) > -1
		})
	}
}

func _buildHostMatcher(host string) func(string) bool {
	switch {
	case host == "*":
		return func(string) bool { return true }

	case host[0] == '*':
		host = host[1:]
		return func(s string) bool { return strings.HasSuffix(s, host) }

	default:
		return func(s string) bool { return s == host }
	}
}

func buildRouteMatcherForHost(m dynamicconfig.Matcher) matcher {
	switch hosts := mergestrings(m.Host, m.Hosts); len(hosts) {
	case 0:
		return nil

	case 1:
		match := _buildHostMatcher(strings.ToLower(hosts[0]))
		return newMatcher(priorityHost, func(c *Context) bool {
			return match(GetHost(c.ClientRequest))
		})

	default:
		matches := make([]func(string) bool, len(hosts))
		for i, host := range hosts {
			matches[i] = _buildHostMatcher(host)
		}
		return newMatcher(priorityHost, func(c *Context) bool {
			host := GetHost(c.ClientRequest)
			for _, match := range matches {
				if match(host) {
					return true
				}
			}
			return false
		})

	}
}

func buildRouteMatcherForHeaders(m dynamicconfig.Matcher) matcher {
	if len(m.Headers) == 0 {
		return nil
	}

	headers := m.Headers
	return newMatcher(priorityHeader, func(c *Context) bool {
		header := c.ClientRequest.Header
		for key, value := range headers {
			values, ok := header[textproto.CanonicalMIMEHeaderKey(key)]
			if !ok || (value != "" && !slices.Contains(values, value)) {
				return false
			}
		}
		return true
	})
}

func buildRouteMatcherForQueries(m dynamicconfig.Matcher) matcher {
	if len(m.Queries) == 0 {
		return nil
	}

	queries := m.Queries
	return newMatcher(priorityQuery, func(c *Context) bool {
		query := c.Queries()
		for key, value := range queries {
			values, ok := query[key]
			if !ok || (value != "" && !slices.Contains(values, value)) {
				return false
			}
		}
		return true
	})
}

func buildRouteMatcherForClientIP(m dynamicconfig.Matcher) (matcher, error) {
	ips := mergestrings(m.ClientIp, m.ClientIps)
	if len(ips) == 0 {
		return nil, nil
	}

	ipchecker, err := nets.NewIPCheckers(ips...)
	if err != nil {
		return nil, err
	}

	return newMatcher(priorityClientIP, func(c *Context) bool {
		return ipchecker.ContainsAddr(GetClientIP(c.ClientRequest))
	}), nil
}

func buildRouteMatcherForServerIP(m dynamicconfig.Matcher) (matcher, error) {
	ips := mergestrings(m.ServerIp, m.ServerIps)
	if len(ips) == 0 {
		return nil, nil
	}

	ipchecker, err := nets.NewIPCheckers(ips...)
	if err != nil {
		return nil, err
	}

	return newMatcher(priorityServerIP, func(c *Context) bool {
		return ipchecker.ContainsAddr(GetServerIP(c.ClientRequest))
	}), nil
}
