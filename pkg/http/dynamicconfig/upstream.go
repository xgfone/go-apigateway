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

package dynamicconfig

import (
	"reflect"
	"slices"
	"sort"
	"time"
)

// Pre-define some upstream hosts.
const (
	HostClient = "$client"
	HostServer = "$server"
)

// Upstream is an upstream configuraiton.
type Upstream struct {
	// Required
	Id        string    `json:"id" yaml:"id"`
	Discovery Discovery `json:"discovery" yaml:"discovery"`

	// Optional
	Timeout time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// Optional
	//
	// lc(least_conn)
	// sh(source_ip_hash, iphash)
	// r(random), wr(weight_random)
	// rr(round_robin), wrr(weight_round_robin),
	//
	// Default: round_robin
	Policy string `json:"policy,omitempty" yaml:"policy,omitempty"`
	Retry  Retry  `json:"retry,omitempty" yaml:"retry,omitempty"`

	// Optional
	Scheme string `json:"scheme,omitempty" yaml:"scheme,omitempty"` // "http(default)", "https"
	Host   string `json:"host,omitempty" yaml:"host,omitempty"`     // "$client"(default), "$server", "xxx"

	// Optional
	Middlewares     Middlewares `json:"middlewares,omitempty" yaml:"middlewares,omitempty"`
	MiddlewareGroup string      `json:"middlewareGroup,omitempty" yaml:"middlewareGroup,omitempty"`
}

// ForwardPolicy returns the normalized forwarding policy.
func (u Upstream) ForwardPolicy() string {
	switch u.Policy {
	case "lc":
		return "least_conn"

	case "sh", "iphash":
		return "source_ip_hash"

	case "r":
		return "random"

	case "wr":
		return "weight_random"

	case "rr", "":
		return "round_robin"

	case "wrr":
		return "weight_round_robin"

	default:
		return u.Policy
	}
}

// Retry is a retry configuration.
type Retry struct {
	Number   int           `json:"number,omitempty" yaml:"number,omitempty"`     // <0: disabled
	Interval time.Duration `json:"interval,omitempty" yaml:"interval,omitempty"` // [0, +∞), Unit: ms
}

// Discovery is the configuration of the upstream server discovery.
type Discovery struct {
	Static *StaticDiscovery `json:"static,omitempty" yaml:"static,omitempty"`
}

// StaticDiscovery is the configuration of the static upstream server discovery.
type StaticDiscovery struct {
	Servers     []Server     `json:"servers,omitempty" yaml:"servers,omitempty"`
	HealthCheck *HealthCheck `json:"healthcheck,omitempty" yaml:"healthcheck,omitempty"`
}

// Server is the configuraiton of an upstream server.
type Server struct {
	Host   string `json:"host,omitempty" yaml:"host,omitempty"`
	Port   uint16 `json:"port,omitempty" yaml:"port,omitempty"`
	Weight int    `json:"weight,omitempty" yaml:"weight,omitempty"`
}

// HealthCheck is the configuration to check the health of upstream servers.
type HealthCheck struct {
	Disable bool     `json:"disable,omitempty" yaml:"disable,omitempty"`
	Checker *Checker `json:"checker,omitempty" yaml:"checker,omitempty"`
	Request *Request `json:"request,omitempty" yaml:"request,omitempty"`

	// FollowRedirect *bool `json:"followRedirect,omitempty" yaml:"followRedirect,omitempty"`
}

// Checker is the configuration of the health checker.
type Checker struct {
	Failure  int           `json:"failure,omitempty" yaml:"failure,omitempty"` // Default: 2
	Timeout  time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Interval time.Duration `json:"interval,omitempty" yaml:"interval,omitempty"` // Default: 10s
}

// Request is the configuration of the request sent by the healthcheck.
type Request struct {
	Method string `json:"method,omitempty" yaml:"method,omitempty"` // default: "GET"
	Scheme string `json:"scheme,omitempty" yaml:"scheme,omitempty"` // "tcp", "http(default)", "https"
	Host   string `json:"host,omitempty" yaml:"host,omitempty"`
	Path   string `json:"path,omitempty" yaml:"path,omitempty"`
	Header Header `json:"header,omitempty" yaml:"header,omitempty"`
}

// Header represents a request header type.
type Header map[string]string

// ------------------------------------------------------------------------ //

// SortUpstreams sorts the upstreams.
func SortUpstreams(ups []Upstream) {
	sort.SliceStable(ups, func(i, j int) bool { return ups[i].Id < ups[j].Id })
}

// DiffUpstreams compares the difference between new and old upstreams,
// and reutrns the added and deleted upstreams.
//
// NOTICE: adds also contains the existed but changed upstreams.
func DiffUpstreams(news, olds []Upstream) (adds, dels []Upstream) {
	SortUpstreams(news)
	SortUpstreams(olds)

	ids := make(map[string]struct{}, len(news))
	adds = make([]Upstream, 0, len(news)/2)
	dels = make([]Upstream, 0, len(olds)/2)

	// add
	for _, up := range news {
		ids[up.Id] = struct{}{}
		index := findupstream(olds, up.Id)
		if index < 0 || !upstreamequal(up, olds[index]) {
			adds = append(adds, up)
		}
	}

	// del
	for _, up := range olds {
		if _, ok := ids[up.Id]; !ok {
			dels = append(dels, up)
		}
	}

	return
}

func upstreamequal(r1, r2 Upstream) bool {
	return reflect.DeepEqual(r1, r2)
}

func findupstream(ups []Upstream, id string) (index int) {
	return slices.IndexFunc(ups, func(r Upstream) bool { return r.Id == id })
}
