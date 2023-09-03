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

// Package discovery provides the building for the upstream discovery.
package discovery

import (
	"github.com/xgfone/go-apigateway/pkg/discovery"
	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
	"github.com/xgfone/go-apigateway/pkg/http/runtime"
)

func init() { runtime.BuildUpstreamDiscovery = buildUpstreamDiscovery }

func buildUpstreamDiscovery(up dynamicconfig.Upstream) (discovery.Discovery, error) {
	return NewStaticDiscovery(up.Id, *up.Discovery.Static)
}
