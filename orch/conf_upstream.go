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
	"reflect"
	"slices"
	"time"
)

// Upstream is an upstream configuraiton.
type Upstream struct {
	// Required
	Id        string    `json:"id" yaml:"id"`
	Discovery Discovery `json:"discovery" yaml:"discovery"`

	// Optional, Default: 0
	Timeout time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// Optional
	Policy string `json:"policy,omitempty" yaml:"policy,omitempty"` // Default: roundrobin
	Retry  Retry  `json:"retry,omitempty" yaml:"retry,omitempty"`

	// Optional
	Scheme string `json:"scheme,omitempty" yaml:"scheme,omitempty"` // "http(default)", "https", "tcp", "tls"
	Host   string `json:"host,omitempty" yaml:"host,omitempty"`     // "$client"(default), "$server", "xxx"
}

// ForwardPolicy returns the normalized forwarding policy.
func (u Upstream) ForwardPolicy() string {
	switch u.Policy {
	case "lc":
		return "leastconn"

	case "sh", "iphash", "hash_sourceip", "hash(sourceip)":
		return "sourceip_hash"

	case "r":
		return "random"

	case "wr":
		return "weight_random"

	case "rr", "":
		return "roundrobin"

	case "wrr":
		return "weight_roundrobin"

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
	Servers []Server `json:"servers,omitempty" yaml:"servers,omitempty"`
}

// Server is the configuraiton of an upstream server.
type Server struct {
	Host   string `json:"host,omitempty" yaml:"host,omitempty"`
	Port   uint16 `json:"port,omitempty" yaml:"port,omitempty"`
	Weight int    `json:"weight,omitempty" yaml:"weight,omitempty"`
}

// ------------------------------------------------------------------------ //

// CompareServer compares the two servers to sort a set of servers.
//
//	-1 if a <  b
//	 0 if a == b
//	 1 if a >  b
func CompareServer(a, b Server) int {
	switch {
	case a.Weight < b.Weight:
		return 1

	case a.Weight > b.Weight:
		return -1

	default:
		switch {
		case a.Host < b.Host:
			return -1

		case a.Host > b.Host:
			return 1

		default:
			return int(a.Port) - int(b.Port)
		}
	}
}

// SortUpstreams sorts the upstreams.
func SortUpstreams(ups []Upstream) {
	slices.SortFunc(ups, func(a, b Upstream) int {
		switch {
		case a.Id < b.Id:
			return -1
		case a.Id == b.Id:
			return 0
		default:
			return 1
		}
	})
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

func upstreamequal(r1, r2 Upstream) bool { return reflect.DeepEqual(r1, r2) }
func findupstream(ups []Upstream, id string) (index int) {
	return slices.IndexFunc(ups, func(r Upstream) bool { return r.Id == id })
}
