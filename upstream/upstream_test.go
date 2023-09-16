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
	"github.com/xgfone/go-loadbalancer/balancer"
	"github.com/xgfone/go-loadbalancer/forwarder"
)

func TestUpstream(t *testing.T) {
	up1 := New(forwarder.New("up1", balancer.DefaultBalancer, loadbalancer.None))

	up1.SetHost("localhost")
	if host := up1.Host(); host != "localhost" {
		t.Errorf("expect host '%s', but got '%s'", "localhost", host)
	}

	up1.SetScheme("https")
	if scheme := up1.Scheme(); scheme != "https" {
		t.Errorf("expect scheme '%s', but got '%s'", "https", scheme)
	}
}
