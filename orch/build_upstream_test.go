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

import "testing"

func TestUpstreamBuild(t *testing.T) {
	up := Upstream{
		Id:      "up1",
		Host:    "localhost",
		Scheme:  "https",
		Timeout: 3,
		Retry: Retry{
			Number:   3,
			Interval: 100,
		},
		Discovery: Discovery{
			Static: &StaticDiscovery{
				Servers: []Server{
					{Host: "127.0.0.1", Port: 8001},
					{Host: "127.0.0.1", Port: 8002},
				},
			},
		},
	}

	if _up, err := up.Build(); err != nil {
		t.Error(err)
	} else if _up == nil {
		t.Errorf("expect an upstream, but got nil")
	}
}
