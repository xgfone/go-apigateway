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

package nets

import (
	"net"
	"testing"
)

func TestIPChecker(t *testing.T) {
	if ip, err := NewIPCheckers("1.2.3.4/16"); err != nil {
		t.Error(err)
	} else if !ip.ContainsString("1.2.5.6") {
		t.Error("expect true, but got false")
	} else if ip.ContainsString("3.4.5.6") {
		t.Error("expect false, but got true")
	} else if !ip.ContainsIP(net.ParseIP("1.2.5.6")) {
		t.Error("expect true, but got false")
	} else if ip.ContainsIP(net.ParseIP("3.4.5.6")) {
		t.Error("expect false, but got true")
	}

	if ip, err := NewIPCheckers("ff00::/128"); err != nil {
		t.Error(err)
	} else if !ip.ContainsString("ff00::") {
		t.Error("expect true, but got false")
	} else if ip.ContainsString("ff00::1") {
		t.Error("expect false, but got true")
	} else if !ip.ContainsIP(net.ParseIP("ff00::")) {
		t.Error("expect true, but got false")
	} else if ip.ContainsIP(net.ParseIP("ff00::1")) {
		t.Error("expect false, but got true")
	}
}
