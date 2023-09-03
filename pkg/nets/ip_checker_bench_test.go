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
	"net/netip"
	"testing"
)

func BenchmarkIPChecker_netipAddr(b *testing.B) {
	ipchecker, err := NewIPChecker("127.0.0.0/8")
	if err != nil {
		panic(err)
	}

	ip, err := netip.ParseAddr("127.0.0.1")
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			ipchecker.ContainsAddr(ip)
		}
	})
}

func BenchmarkIPChecker_netIP(b *testing.B) {
	ipchecker, err := NewIPChecker("127.0.0.0/8")
	if err != nil {
		panic(err)
	}

	ip := net.ParseIP("127.0.0.1")
	if ip == nil {
		panic("ip is nil")
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			ipchecker.ContainsIP(ip)
		}
	})
}

func BenchmarkIPChecker_string(b *testing.B) {
	ipchecker, err := NewIPChecker("127.0.0.0/8")
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			ipchecker.ContainsString("127.0.0.1")
		}
	})
}
