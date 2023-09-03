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
	"strings"
)

// IPChecker is used to check whether an ip is contained.
type IPChecker struct{ netip.Prefix }

// NewIPChecker returns a new ip checker.
func NewIPChecker(cidr string) (c IPChecker, err error) {
	c.Prefix, err = netip.ParsePrefix(cidr)
	return
}

// String returns the description of the ip checker.
func (c IPChecker) String() string { return c.Prefix.String() }

// ContainsIP reports whether the checker contains the ip.
func (c IPChecker) ContainsIP(ip net.IP) bool {
	if len(ip) == 0 {
		return false
	}

	if c.Prefix.Addr().BitLen() == 32 {
		ip = ip.To4()
	} else {
		ip = ip.To16()
	}

	addr, ok := netip.AddrFromSlice(ip)
	return ok && c.Prefix.Contains(addr)
}

// ContainsAddr reports whether the checker contains the ip addr.
func (c IPChecker) ContainsAddr(ip netip.Addr) bool {
	return c.Prefix.Contains(ip)
}

// ContainsString reports whether the checker contains the ip string.
func (c IPChecker) ContainsString(ip string) bool {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}
	return c.Prefix.Contains(addr)
}

// ------------------------------------------------------------------------ //

// IPCheckers is a set of ip checkers.
type IPCheckers []IPChecker

// NewIPCheckers returns a new ip checkers.
func NewIPCheckers(cidrs ...string) (cs IPCheckers, err error) {
	cs = make(IPCheckers, len(cidrs))
	for i, cidr := range cidrs {
		cs[i], err = NewIPChecker(cidr)
		if err != nil {
			return
		}
	}
	return
}

// String returns the description of the ip checkers.
func (cs IPCheckers) String() string {
	ss := make([]string, len(cs))
	for i, s := range cs {
		ss[i] = s.String()
	}
	return strings.Join(ss, ",")
}

// ContainsIP reports whether the checkers contains the ip.
func (cs IPCheckers) ContainsIP(ip net.IP) bool {
	for _, c := range cs {
		if c.ContainsIP(ip) {
			return true
		}
	}
	return false
}

// ContainsAddr reports whether the checkers contains the ip addr.
func (cs IPCheckers) ContainsAddr(ip netip.Addr) bool {
	for _, c := range cs {
		if c.ContainsAddr(ip) {
			return true
		}
	}
	return false
}

// ContainsString reports whether the checkers contains the ip string.
func (cs IPCheckers) ContainsString(ip string) bool {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}
	return cs.ContainsAddr(addr)
}
