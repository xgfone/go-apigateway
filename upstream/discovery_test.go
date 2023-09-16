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

package upstream

import (
	"testing"

	"github.com/xgfone/go-loadbalancer"
	"github.com/xgfone/go-loadbalancer/endpoint"
)

func TestNewDiscovery(t *testing.T) {
	if d := NewDiscovery().Discover(); d != loadbalancer.None {
		t.Error("expect none static endpoints, but got other")
	} else if d.Endpoints != nil {
		t.Errorf("expect endpoints nil, but got %+v", d.Endpoints)
	}

	ep := endpoint.New("id", nil)
	if d := NewDiscovery(ep).Discover(); d == nil {
		t.Error("unexpect static endpoints is nil")
	} else if len(d.Endpoints) != 1 {
		t.Errorf("expect 1 endpoint, but got %d", len(d.Endpoints))
	} else if id := d.Endpoints[0].ID(); id != "id" {
		t.Errorf("expect endpoint '%s', but got '%s'", "id", id)
	}
}
