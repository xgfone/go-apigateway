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

	matcher "github.com/xgfone/go-http-matcher"
)

var noroutematches = errors.New("matcher exists, but has no any matching items")

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

func (m *HttpMatcher) build() ([]matcher.Matcher, error) {
	ms := make([]matcher.Matcher, 0, 4)

	ms = _appendmatcher(ms, matcher.Host(m.Hosts...))
	ms = _appendmatcher(ms, matcher.Method(m.Methods...))
	ms = _appendmatcher(ms, matcher.Headerm(m.Headers))
	ms = _appendmatcher(ms, matcher.Querym(m.Queries))
	ms = _appendmatcher(ms, matcher.PathPrefix(m.PathPrefixes...))
	ms = _appendmatcher(ms, matcher.Path(m.Paths...))

	clientIPMatcher, err := matcher.ClientIP(m.ClientIps...)
	if err != nil {
		return nil, err
	}
	ms = _appendmatcher(ms, clientIPMatcher)

	serverIPMatcher, err := matcher.ServerIP(m.ServerIps...)
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
