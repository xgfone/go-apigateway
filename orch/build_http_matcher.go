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
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/matcher"
	"github.com/xgfone/go-apigateway/nets"
)

// Build builds the matcher.
func (m HttpMatcher) Build() (matcher.Matcher, error) {
	ms, err := m.build()
	if err != nil {
		return nil, err
	} else if len(ms) == 0 {
		return nil, noroutematches
	}
	return matcher.And(ms...), nil
}

// Build builds a set of matchers to a route matcher.
func (ms HttpMatchers) Build() (m matcher.Matcher, err error) {
	_ms := make([]matcher.Matcher, len(ms))
	for i, _m := range ms {
		if _ms[i], err = _m.Build(); err != nil {
			return
		}
	}
	return matcher.Or(_ms...), nil
}

// ----------------------------------------------------------------------- //

var noroutematches = errors.New("matcher exists, but has no any matching items")

func (m *HttpMatcher) build() ([]matcher.Matcher, error) {
	ms := make([]matcher.Matcher, 0, 4)

	ms = _appendmatcher(ms, m.buildHost())
	ms = _appendmatcher(ms, m.buildMethod())
	ms = _appendmatcher(ms, m.buildHeaders())
	ms = _appendmatcher(ms, m.buildQueries())
	ms = _appendmatcher(ms, m.buildPathPrefix())
	ms = _appendmatcher(ms, m.buildPath())

	clientIPMatcher, err := m.buildClientIP()
	if err != nil {
		return nil, err
	}
	ms = _appendmatcher(ms, clientIPMatcher)

	serverIPMatcher, err := m.buildServerIP()
	if err != nil {
		return nil, err
	}
	ms = _appendmatcher(ms, serverIPMatcher)

	return ms, nil
}

func _appendmatcher(ms []matcher.Matcher, m matcher.Matcher) []matcher.Matcher {
	if m != nil {
		ms = append(ms, m)
	}
	return ms
}

const (
	priorityQuery      = 1
	priorityHeader     = 4
	priorityClientIP   = 20
	priorityServerIP   = 20
	priorityMethod     = 40
	priorityPathPrefix = 50
	priorityPath       = 500
	priorityHost       = 5000
)

// var defaultPathPrefixMatcher = newPathPrefixMatcher("/")

func fixPath(path string) string {
	if path == "/" {
		return path
	} else if path = strings.TrimRight(path, "/"); path != "" {
		return path
	}
	return "/"
}

type (
	exactFullMatches   []string
	exactPrefixMatches []string
)

func (ps exactFullMatches) Match(path string) bool {
	for _, s := range ps {
		if path == s {
			return true
		}
	}
	return false
}

func (ps exactPrefixMatches) Match(path string) bool {
	for _, s := range ps {
		if strings.HasPrefix(path, s) {
			return true
		}
	}
	return false
}

func (m *HttpMatcher) buildPath() matcher.Matcher {
	switch _len := len(m.Paths); _len {
	case 0:
		return nil

	case 1:
		path := fixPath(m.Paths[0])
		desc := fmt.Sprintf("Path(`%s`)", path)
		return matcher.New(priorityPath*len(path), desc, func(c *core.Context) bool {
			return matcher.GetPath(c.ClientRequest) == path
		})
	}

	var maxlen int
	paths := make(exactFullMatches, len(m.Paths))
	for i, path := range m.Paths {
		paths[i] = fixPath(path)
		if _len := len(paths[i]); _len > maxlen {
			maxlen = _len
		}
	}

	desc := fmt.Sprintf("Path(`%s`)", strings.Join(paths, "`,`"))
	return matcher.New(priorityPath*maxlen, desc, func(c *core.Context) bool {
		return paths.Match(matcher.GetPath(c.ClientRequest))
	})
}

func newPathPrefixMatcher(prefix string) matcher.Matcher {
	desc := fmt.Sprintf("PathPrefix(`%s`)", prefix)
	return matcher.New(priorityPathPrefix*len(prefix), desc, func(c *core.Context) bool {
		return strings.HasPrefix(matcher.GetPath(c.ClientRequest), prefix)
	})
}

func (m *HttpMatcher) buildPathPrefix() matcher.Matcher {
	switch _len := len(m.PathPrefixes); _len {
	case 0:
		return nil

	case 1:
		return newPathPrefixMatcher(fixPath(m.PathPrefixes[0]))
	}

	var maxlen int
	prefixs := make(exactPrefixMatches, len(m.PathPrefixes))
	for i, prefix := range m.PathPrefixes {
		prefixs[i] = fixPath(prefix)
		if _len := len(prefixs[i]); _len > maxlen {
			maxlen = _len
		}
	}

	desc := fmt.Sprintf("PathPrefix(`%s`)", strings.Join(prefixs, "`,`"))
	return matcher.New(priorityPathPrefix*maxlen, desc, func(c *core.Context) bool {
		return prefixs.Match(matcher.GetPath(c.ClientRequest))
	})
}

func (m *HttpMatcher) buildMethod() matcher.Matcher {
	switch _len := len(m.Methods); _len {
	case 0:
		return nil

	case 1:
		method := strings.ToUpper(m.Methods[0])
		desc := fmt.Sprintf("Method(`%s`)", method)
		return matcher.New(priorityMethod, desc, func(c *core.Context) bool {
			return c.ClientRequest.Method == method
		})
	}

	methods := make(exactFullMatches, len(m.Methods))
	for i, method := range m.Methods {
		methods[i] = strings.ToUpper(method)
	}

	desc := fmt.Sprintf("Method(`%s`)", strings.Join(methods, "`,`"))
	return matcher.New(priorityMethod, desc, func(c *core.Context) bool {
		return methods.Match(c.ClientRequest.Method)
	})
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

func (m *HttpMatcher) buildHost() matcher.Matcher {
	switch _len := len(m.Hosts); _len {
	case 0:
		return nil

	case 1:
		host := strings.ToLower(m.Hosts[0])
		desc := fmt.Sprintf("Host(`%s`)", host)
		match := _buildHostMatcher(host)
		return matcher.New(priorityHost*len(host), desc, func(c *core.Context) bool {
			return match(matcher.GetHost(c.ClientRequest))
		})
	}

	var maxlen int
	matches := make([]func(string) bool, len(m.Hosts))
	for i, host := range m.Hosts {
		host = strings.ToLower(host)
		matches[i] = _buildHostMatcher(host)
		if _len := len(host); _len > maxlen {
			maxlen = _len
		}
	}

	desc := fmt.Sprintf("Host(`%s`)", strings.Join(m.Hosts, "`,`"))
	return matcher.New(priorityHost*maxlen, desc, func(c *core.Context) bool {
		host := matcher.GetHost(c.ClientRequest)
		for _, match := range matches {
			if match(host) {
				return true
			}
		}
		return false
	})
}

func (m *HttpMatcher) buildHeaders() matcher.Matcher {
	if len(m.Headers) == 0 {
		return nil
	}

	headers := make(map[string]string, len(m.Headers))
	for key, value := range m.Headers {
		headers[http.CanonicalHeaderKey(key)] = value
	}

	// TODO:
	desc := ""

	return matcher.New(priorityHeader*len(headers), desc, func(c *core.Context) bool {
		header := c.ClientRequest.Header
		for key, value := range headers {
			values, ok := header[key]
			if !ok || (value != "" && !slices.Contains(values, value)) {
				return false
			}
		}
		return true
	})
}

func (m *HttpMatcher) buildQueries() matcher.Matcher {
	if len(m.Queries) == 0 {
		return nil
	}

	// TODO:
	desc := ""

	queries := m.Queries
	return matcher.New(priorityQuery*len(queries), desc, func(c *core.Context) bool {
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

func (m *HttpMatcher) buildClientIP() (matcher.Matcher, error) {
	if len(m.ClientIps) == 0 {
		return nil, nil
	}

	ipchecker, err := nets.NewIPCheckers(m.ClientIps...)
	if err != nil {
		return nil, err
	}

	desc := fmt.Sprintf("ClientIps(`%s`)", strings.Join(m.ClientIps, "`,`"))
	return matcher.New(priorityClientIP, desc, func(c *core.Context) bool {
		return ipchecker.ContainsAddr(matcher.GetClientIP(c.ClientRequest))
	}), nil
}

func (m *HttpMatcher) buildServerIP() (matcher.Matcher, error) {
	if len(m.ServerIps) == 0 {
		return nil, nil
	}

	ipchecker, err := nets.NewIPCheckers(m.ServerIps...)
	if err != nil {
		return nil, err
	}

	desc := fmt.Sprintf("ServerIps(`%s`)", strings.Join(m.ServerIps, "`,`"))
	return matcher.New(priorityServerIP, desc, func(c *core.Context) bool {
		return ipchecker.ContainsAddr(matcher.GetServerIP(c.ClientRequest))
	}), nil
}
