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
)

func TestDiffUpstreams(t *testing.T) {
	ups1 := []Upstream{
		{
			Id: "up1",
			Discovery: Discovery{
				Static: &StaticDiscovery{
					Servers: []Server{
						{Host: "127.0.0.1", Port: 8001},
						{Host: "127.0.0.1", Port: 8002},
					},
				},
			},
		},

		{
			Id: "up2",
			Discovery: Discovery{
				Static: &StaticDiscovery{
					Servers: []Server{
						{Host: "127.0.0.1", Port: 8003},
						{Host: "127.0.0.1", Port: 8004},
					},
				},
			},
		},
	}

	ups2 := []Upstream{
		{
			Id: "up1",
			Discovery: Discovery{
				Static: &StaticDiscovery{
					Servers: []Server{
						{Host: "127.0.0.1", Port: 8001},
						{Host: "127.0.0.1", Port: 8002},
					},
				},
			},
		},

		{
			Id: "up2",
			Discovery: Discovery{
				Static: &StaticDiscovery{
					Servers: []Server{
						{Host: "127.0.0.1", Port: 8003},
						{Host: "127.0.0.1", Port: 8004},
					},
				},
			},
		},
	}

	adds, dels := DiffUpstreams(ups1, ups2)
	if len(adds) != 0 {
		t.Errorf("unexpect added upstreams: %+v", adds)
	}
	if len(dels) != 0 {
		t.Errorf("unexpect deleted upstreams: %+v", dels)
	}

	ups1[1].Id = "up3"
	adds, dels = DiffUpstreams(ups1, ups2)
	if len(adds) != 1 {
		t.Errorf("expect %d added upstreams, but got %d: %+v", 1, len(adds), adds)
	} else if adds[0].Id != "up3" {
		t.Errorf("expect added upstream '%s', but got '%s'", "up3", adds[0].Id)
	}

	if len(dels) != 1 {
		t.Errorf("expect %d deleted upstreams, but got %d: %+v", 1, len(dels), dels)
	} else if dels[0].Id != "up2" {
		t.Errorf("expect deleted upstream '%s', but got '%s'", "up2", dels[0].Id)
	}
}

func TestCompareServer(t *testing.T) {
	servers := []Server{
		{Host: "127.0.0.1", Port: 80, Weight: 1},
		{Host: "127.0.0.1", Port: 80, Weight: 2},
		{Host: "127.0.0.1", Port: 80, Weight: 3},
		{Host: "127.0.0.2", Port: 80, Weight: 1},
		{Host: "127.0.0.1", Port: 10, Weight: 1},
	}
	slices.SortFunc(servers, CompareServer)

	expects := []Server{
		{Host: "127.0.0.1", Port: 80, Weight: 3},
		{Host: "127.0.0.1", Port: 80, Weight: 2},
		{Host: "127.0.0.1", Port: 10, Weight: 1},
		{Host: "127.0.0.1", Port: 80, Weight: 1},
		{Host: "127.0.0.2", Port: 80, Weight: 1},
	}

	if !reflect.DeepEqual(servers, expects) {
		t.Errorf("expect %+v, but got %+v", expects, servers)
	}
}
