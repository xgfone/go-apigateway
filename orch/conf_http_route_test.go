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
	"testing"
	"time"
)

var (
	httproute1 = HttpRoute{
		Id:       "route1",
		Upstream: "upstream",

		Priority:        123,
		RequestTimeout:  time.Second,
		MiddlewareGroup: "group",
		Middlewares: Middlewares{
			{Name: "allow", Conf: "127.0.0.1/8"},
			{Name: "block", Conf: map[string]any{"cidrs": []string{"0.0.0.0/0"}}},
		},

		Matchers: []HttpMatcher{
			{
				Methods: []string{"GET"},
				Paths:   []string{"/path1"},
				Headers: map[string]string{"key1": "value1", "key2": "value2"},
			},
			{
				Methods: []string{"GET"},
				Paths:   []string{"/path2"},
				Headers: map[string]string{"key1": "value1", "key2": "value2"},
			},
		},
	}

	httproute2 = HttpRoute{
		Id:       "route2",
		Upstream: "upstream",

		Priority:        123,
		RequestTimeout:  time.Second,
		MiddlewareGroup: "group",
		Middlewares: Middlewares{
			{Name: "allow", Conf: "127.0.0.1/8"},
			{Name: "block", Conf: map[string]any{"cidrs": []string{"0.0.0.0/0"}}},
		},

		Matchers: []HttpMatcher{
			{
				Methods: []string{"GET"},
				Paths:   []string{"/path1"},
				Headers: map[string]string{"key2": "value2", "key1": "value1"},
			},
			{
				Methods: []string{"GET"},
				Paths:   []string{"/path2"},
				Headers: map[string]string{"key2": "value2", "key1": "value1"},
			},
		},
	}
)

func TestRouteEqual(t *testing.T) {
	r1 := HttpRoute{Matchers: []HttpMatcher{{Paths: []string{"/path1"}}, {Paths: []string{"/path2"}}}}
	r2 := HttpRoute{Matchers: []HttpMatcher{{Paths: []string{"/path2"}}, {Paths: []string{"/path1"}}}}
	if routeequal(r1, r2) {
		t.Errorf("unexpect two routes are equal")
	}
}

func TestDiffRoutes(t *testing.T) {
	route1 := httproute1
	route2 := httproute2
	routes1 := []HttpRoute{route1, route2}
	routes2 := []HttpRoute{route2, route1}

	adds, dels := DiffHttpRoutes(routes1, routes2)
	if len(adds) != 0 {
		t.Errorf("unexpect added routes, but got %+v", adds)
	}
	if len(dels) != 0 {
		t.Errorf("unexpect deled routes, but got %+v", dels)
	}

	routes1[0].Matchers = slices.Clone(routes1[0].Matchers)
	routes1[0].Matchers[0].Paths = []string{"/path3"}
	adds, dels = DiffHttpRoutes(routes1, routes2)
	if len(adds) != 1 {
		t.Errorf("expect one added routes, but got %d: %+v", len(adds), adds)
	} else if !routeequal(adds[0], routes1[0]) {
		t.Errorf("expect added route %+v, but got %+v", routes1[0], adds[0])
	}
	if len(dels) != 0 {
		t.Errorf("unexpect deled routes, but got %d: %+v", len(dels), dels)
	}

	routes2[1].Id = "route3"
	adds, dels = DiffHttpRoutes(routes1, routes2)
	if !reflect.DeepEqual(adds, routes1[:1]) {
		t.Errorf("expect added routes %+v, but got %+v", routes1[:1], adds)
	}
	if !reflect.DeepEqual(dels, routes2[1:]) {
		t.Errorf("expect deled routes %+v, but got %+v", routes2[1:], dels)
	}
}
