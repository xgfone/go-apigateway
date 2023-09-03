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

// Route is a route configuration.
type Route struct {
	// Required
	Id       string  `json:"id" yaml:"id"`
	Matcher  Matcher `json:"matcher" yaml:"matcher"`
	Upstream string  `json:"upstream" yaml:"upstream"`

	// Optional
	Protect         bool          `json:"protect,omitempty" yaml:"protect,omitempty"`
	Priority        int           `json:"priority,omitempty" yaml:"priority,omitempty"`
	Timeout         time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Middlewares     Middlewares   `json:"middlewares,omitempty" yaml:"middlewares,omitempty"`
	MiddlewareGroup string        `json:"middlewareGroup,omitempty" yaml:"middlewareGroup,omitempty"`
}

// Matcher is the configuraiton of a route matcher.
type Matcher struct {
	Method  string   `json:"method,omitempty" yaml:"method,omitempty"`
	Methods []string `json:"methods,omitempty" yaml:"methods,omitempty"`

	// Without Argument
	Path  string   `json:"path,omitempty" yaml:"path,omitempty"`
	Paths []string `json:"paths,omitempty" yaml:"paths,omitempty"`

	// Without Argument
	PathPrefix  string   `json:"pathPrefix,omitempty" yaml:"pathPrefix,omitempty"`
	PathPrefixs []string `json:"pathPrefixs,omitempty" yaml:"pathPrefixs,omitempty"`

	// Exact or wildcard
	Host  string   `json:"host,omitempty" yaml:"host,omitempty"`
	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`

	// Exact
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Queries map[string]string `json:"queries,omitempty" yaml:"queries,omitempty"`

	// IPv4CIDR or IPv6CIDR
	ClientIp  string   `json:"clientIp,omitempty" yaml:"clientIp,omitempty"`
	ClientIps []string `json:"clientIps,omitempty" yaml:"clientIps,omitempty"`

	// IPv4CIDR or IPv6CIDR
	ServerIp  string   `json:"serverIp,omitempty" yaml:"serverIp,omitempty"`
	ServerIps []string `json:"serverIps,omitempty" yaml:"serverIps,omitempty"`
}

// ------------------------------------------------------------------------ //

// SortRoutes sorts the routes.
func SortRoutes(routes []Route) {
	sort.SliceStable(routes, func(i, j int) bool { return routes[i].Id < routes[j].Id })
}

// DiffRoutes compares the difference between new and old routes,
// and reutrns the added and deleted routes.
func DiffRoutes(news, olds []Route) (adds, dels []Route) {
	SortRoutes(news)
	SortRoutes(olds)

	ids := make(map[string]struct{}, len(news))
	adds = make([]Route, 0, len(news)/2)
	dels = make([]Route, 0, len(olds)/2)

	// add
	for _, route := range news {
		ids[route.Id] = struct{}{}
		index := findroute(olds, route.Id)
		if index < 0 || !routeequal(route, olds[index]) {
			adds = append(adds, route)
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

func routeequal(r1, r2 Route) bool {
	return reflect.DeepEqual(r1, r2)
}

func findroute(routes []Route, id string) (index int) {
	return slices.IndexFunc(routes, func(r Route) bool { return r.Id == id })
}
