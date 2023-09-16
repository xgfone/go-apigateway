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
	"strings"
	"testing"

	"github.com/xgfone/go-apigateway/http/middleware"
	_ "github.com/xgfone/go-apigateway/http/middleware/middlewares"
)

func buildHttpMiddlewareGroups(gs []MiddlewareGroup) ([]*middleware.Group, error) {
	var err error
	_gs := make([]*middleware.Group, len(gs))
	for i, g := range gs {
		if _gs[i], err = g.HttpBuild(); err != nil {
			return nil, err
		}
	}
	return _gs, nil
}

func TestBuildHTTPMiddlewareGroups(t *testing.T) {
	groups := []MiddlewareGroup{
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

	gs, err := buildHttpMiddlewareGroups(groups)
	if err != nil {
		t.Fatal(err)
	}

	if len(gs) != 2 {
		t.Errorf("expect %d middleware groups, but got %d: %+v", 2, len(gs), gs)
	} else {
		for _, g := range gs {
			switch g.Name() {
			case "g1", "g2":
			default:
				t.Errorf("expect middleware group '%s'", g.Name())
			}

			if len(g.Middlewares()) != 2 {
				t.Errorf("Group<%s>: expect %d middlewares, but got %d", g.Name(), 2, len(g.Middlewares()))
			}

			for _, m := range g.Middlewares() {
				switch m.Name() {
				case "allow", "block":
				default:
					t.Errorf("Group<%s>: unexpect middleware '%s'", g.Name(), m.Name())
				}
			}
		}
	}

	groups[0].Name = ""
	if _, err := buildHttpMiddlewareGroups(groups); err == nil {
		t.Errorf("expect an error, but not nil")
	} else if s := err.Error(); !strings.HasPrefix(s, "HttpMiddlewareGroup: ") {
		t.Errorf("unexpect error '%s'", s)
	}

	groups[0].Name = "g1"
	groups[0].Middlewares[0].Name = ""
	if _, err := buildHttpMiddlewareGroups(groups); err == nil {
		t.Errorf("expect an error, but not nil")
	} else if s := err.Error(); !strings.Contains(s, "HttpMiddleware: ") {
		t.Errorf("unexpect error '%s'", s)
	}
}
