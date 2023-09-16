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

package router

import (
	"time"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/upstream"
)

// AfterRoute is the next handler after the route.
var AfterRoute core.Handler = upstream.Forward

var (
	// AlwaysTrue is a route matcher that always reutrns true.
	AlwaysTrue = MatcherFunc(func(c *core.Context) bool { return true })

	// AlwaysFalse is a route matcher that always reutrns false.
	AlwaysFalse = MatcherFunc(func(c *core.Context) bool { return false })
)

// Matcher is used to check whether the rule matches the request.
type Matcher interface {
	Match(*core.Context) bool
}

// MatcherFunc is a route matcher function.
type MatcherFunc func(c *core.Context) bool

// Match implements the interface Matcher.
func (f MatcherFunc) Match(c *core.Context) bool { return f(c) }

// Route represents a runtime route.
type Route struct {
	// Required
	RouteId    string `json:"routeId" yaml:"routeId"`
	UpstreamId string `json:"upstreamId" yaml:"upstreamId"`

	// Optional
	//
	// The bigger the value, the higher the priority.
	Priority int `json:"priority,omitempty" yaml:"priority,omitempty"`

	// Optional
	//
	// If true, the route is called in the apigateway inside, and not routed.
	Protect bool `json:"protect,omitempty" yaml:"protect,omitempty"`

	// Optional
	RequestTimeout time.Duration `json:"requestTimeout,omitempty" yaml:"requestTimeout,omitempty"`
	ForwardTimeout time.Duration `json:"forwardTimeout,omitempty" yaml:"forwardTimeout,omitempty"`

	// Optional
	//
	// The original configuration of the route.
	Config interface{} `json:"config,omitempty" yaml:"config,omitempty"`

	// Optional
	//
	// It may be the description of the matcher.
	Desc string `json:"desc" yaml:"desc"`

	Matcher        `json:"-" yaml:"-"` // Required
	core.Handler   `json:"-" yaml:"-"` // Optional, Default: AfterRoute
	core.Responser `json:"-" yaml:"-"` // Optional, Default: core.StdResponse
}
