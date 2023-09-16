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
	"testing"
)

func TestDiffMiddlewareGroup(t *testing.T) {
	group1 := []MiddlewareGroup{
		{
			Name: "g1",
			Middlewares: Middlewares{
				{
					Name: "allow",
					Conf: []string{"127.0.0.0/8"},
				},
				{
					Name: "block",
					Conf: []string{"0.0.0.0/0"},
				},
			},
		},

		{
			Name: "g2",
			Middlewares: Middlewares{
				{
					Name: "allow",
					Conf: []string{"127.0.0.0/8"},
				},
				{
					Name: "block",
					Conf: []string{"0.0.0.0/0"},
				},
			},
		},
	}

	group2 := []MiddlewareGroup{
		{
			Name: "g1",
			Middlewares: Middlewares{
				{
					Name: "allow",
					Conf: []string{"127.0.0.0/8"},
				},
				{
					Name: "block",
					Conf: []string{"0.0.0.0/0"},
				},
			},
		},

		{
			Name: "g2",
			Middlewares: Middlewares{
				{
					Name: "allow",
					Conf: []string{"127.0.0.0/8"},
				},
				{
					Name: "block",
					Conf: []string{"0.0.0.0/0"},
				},
			},
		},
	}

	adds, dels := DiffMiddlewareGroups(group1, group2)
	if len(adds) != 0 {
		t.Errorf("unexpect added middleware group, but got %d: %+v", len(adds), adds)
	}
	if len(dels) != 0 {
		t.Errorf("unexpect deleted middleware group, but got %d: %+v", len(dels), dels)
	}

	oldmws := group2[1].Middlewares
	group2[1].Middlewares = Middlewares{
		{
			Name: "block",
			Conf: []string{"0.0.0.0/0"},
		},
		{
			Name: "redirect",
			Conf: map[string]any{"code": 302, "location": "http://localhost"},
		},
	}

	adds, dels = DiffMiddlewareGroups(group1, group2)
	if len(adds) != 1 {
		t.Errorf("expect one added middlewares, but got %d: %+v", len(adds), adds)
	} else if !reflect.DeepEqual(adds[0], group1[1]) {
		t.Errorf("expect added middleware %+v, but got %+v", group1[1], adds[0])
	}
	if len(dels) != 0 {
		t.Errorf("unexpect deleted middleware groups, but got %d: %+v", len(dels), dels)
	}

	group2[1].Name = "g3"
	adds, dels = DiffMiddlewareGroups(group1, group2)
	if len(adds) != 1 {
		t.Errorf("expect one added middlewares, but got %d: %+v", len(adds), adds)
	} else if !reflect.DeepEqual(adds[0], group1[1]) {
		t.Errorf("expect added middleware %+v, but got %+v", group1[1], adds[0])
	} else if !reflect.DeepEqual(adds[0].Middlewares, oldmws) {
		t.Errorf("expect added middlewares %+v, but got %+v", oldmws, dels[0].Middlewares)
	}

	if len(dels) != 1 {
		t.Errorf("expect one deleted middleware group, but got %d: %+v", len(dels), dels)
	} else if !reflect.DeepEqual(dels[0], group2[1]) {
		t.Errorf("expect deleted middleware group %+v, but got %+v", group2[1], dels[0])
	}
}
