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
	"fmt"
	"reflect"
	"slices"
	"time"
)

// HttpRoute is a http route configuration.
type HttpRoute struct {
	// Required
	Id       string       `json:"id" yaml:"id"`
	Upstream string       `json:"upstream" yaml:"upstream"`
	Matchers HttpMatchers `json:"matchers" yaml:"matchers"`

	// Optional
	Protect  bool `json:"protect,omitempty" yaml:"protect,omitempty"`
	Priority int  `json:"priority,omitempty" yaml:"priority,omitempty"`

	RequestTimeout time.Duration `json:"requestTimeout,omitempty" yaml:"requestTimeout,omitempty"`
	ForwardTimeout time.Duration `json:"forwardTimeout,omitempty" yaml:"forwardTimeout,omitempty"`

	Middlewares      Middlewares `json:"middlewares,omitempty" yaml:"middlewares,omitempty"`
	MiddlewareGroups []string    `json:"middlewareGroups,omitempty" yaml:"middlewareGroups,omitempty"`
}

// HttpRouteMatcher is the configuraiton of a route matcher.
type HttpMatcher struct {
	// Exact(www.example.com) or Wildcard(*.example.com)
	//
	// OR Match
	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`

	// Exact OR Match
	Methods []string `json:"methods,omitempty" yaml:"methods,omitempty"`

	// Exact OR Match
	Paths        []string `json:"paths,omitempty" yaml:"paths,omitempty"`
	PathPrefixes []string `json:"pathPrefixes,omitempty" yaml:"pathPrefixes,omitempty"`

	// Exact Match
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Queries map[string]string `json:"queries,omitempty" yaml:"queries,omitempty"`

	// IPv4CIDR or IPv6CIDR
	ClientIps []string `json:"clientIps,omitempty" yaml:"clientIps,omitempty"`
	ServerIps []string `json:"serverIps,omitempty" yaml:"serverIps,omitempty"`
}

// HttpMatchers represents a set of http matchers.
type HttpMatchers []HttpMatcher

// ------------------------------------------------------------------------ //

// DiffHttpRoutes compares the difference between new and old routes,
// and reutrns the added and deleted routes.
func DiffHttpRoutes(news, olds []HttpRoute) (adds, dels []HttpRoute) {
	ids := make(map[string]struct{}, len(news))
	adds = make([]HttpRoute, 0, len(news)/2)
	dels = make([]HttpRoute, 0, len(olds)/2)

	// add
	for _, route := range news {
		ids[route.Id] = struct{}{}
		index := findroute(olds, route.Id)
		if index < 0 || !routeequal(route, olds[index]) {
			adds = append(adds, route)
		} else {
			fmt.Println(index, !routeequal(route, olds[index]))
			fmt.Println(route, olds[index])
		}
	}

	// del
	for _, route := range olds {
		if _, ok := ids[route.Id]; !ok {
			dels = append(dels, route)
		}
	}

	return
}

func routeequal(r1, r2 HttpRoute) bool { return reflect.DeepEqual(r1, r2) }
func findroute(routes []HttpRoute, id string) (index int) {
	return slices.IndexFunc(routes, func(r HttpRoute) bool { return r.Id == id })
}
