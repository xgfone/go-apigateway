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

// Package provider defines a group of the runtime configuration providers.
package provider

import "github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"

type (
	// RouteProvider is the provider for the routes.
	RouteProvider interface {
		Routes(lastEtag string) (routes []dynamicconfig.Route, etag string, err error)
	}

	// UpstreamProvider is the provider for the upstreams.
	UpstreamProvider interface {
		Upstreams(lastEtag string) (upstreams []dynamicconfig.Upstream, etag string, err error)
	}

	// MiddlewareGroupProvider is the provider for the middleware groups.
	MiddlewareGroupProvider interface {
		MiddlewareGroups(lastEtag string) (groups dynamicconfig.MiddlewareGroups, etag string, err error)
	}
)

type (
	// RouteProviderFunc is a RouteProvider function.
	RouteProviderFunc func(string) ([]dynamicconfig.Route, string, error)

	// UpstreamProviderFunc is a UpstreamProvider function.
	UpstreamProviderFunc func(string) ([]dynamicconfig.Upstream, string, error)

	// MiddlewareGroupProviderFunc is a MiddlewareGroupProvider function.
	MiddlewareGroupProviderFunc func(string) (dynamicconfig.MiddlewareGroups, string, error)
)

// Routes implements the interface RouteProvider.
func (f RouteProviderFunc) Routes(etag string) ([]dynamicconfig.Route, string, error) {
	return f(etag)
}

// Upstreams implements the interface UpstreamProvider.
func (f UpstreamProviderFunc) Upstreams(etag string) ([]dynamicconfig.Upstream, string, error) {
	return f(etag)
}

// MiddlewareGroups implements the interface MiddlewareGroupProvider.
func (f MiddlewareGroupProviderFunc) MiddlewareGroups(etag string) (dynamicconfig.MiddlewareGroups, string, error) {
	return f(etag)
}
