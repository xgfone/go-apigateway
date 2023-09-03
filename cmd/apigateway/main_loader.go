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

package main

import (
	"flag"
	"time"

	_ "github.com/xgfone/go-apigateway/pkg/http/discovery"
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares"

	"github.com/xgfone/go-apigateway/pkg/http/loader"
	"github.com/xgfone/go-apigateway/pkg/http/provider/localfiles"
	"github.com/xgfone/go-atexit"
)

var (
	_                 = flag.String("provider", "localfiledir", "The provider of the dynamic configurations.")
	routeslocaldir    = flag.String("provider.localfiledir.routes", "routes", "The directory of the local files storing the dynamic routes.")
	upstreamslocaldir = flag.String("provider.localfiledir.upstreams", "upstreams", "The directory of the local files storing the dynamic upstreams.")
	mdwgroupslocaldir = flag.String("provider.localfiledir.middlewaregroups", "middlewaregroups", "The directory of the local files storing the dynamic middleware groups.")
)

var (
	routesLoader    loader.ResourceLoader
	upstreamsLoader loader.ResourceLoader
	mdwgroupsLoader loader.ResourceLoader
)

func initloader() {
	routesProvider := localfiles.RouteProvider(*routeslocaldir)
	upstreamsProvider := localfiles.UpstreamProvider(*upstreamslocaldir)
	mdwgroupsProvider := localfiles.MiddlewareGroupProvider(*mdwgroupslocaldir)

	routesLoader = loader.RouteProviderLoader(routesProvider, time.Minute)
	upstreamsLoader = loader.UpstreamProviderLoader(upstreamsProvider, time.Minute)
	mdwgroupsLoader = loader.MiddlewareGroupProviderLoader(mdwgroupsProvider, time.Minute)

	go routesLoader.Load(atexit.Context())
	go upstreamsLoader.Load(atexit.Context())
	go mdwgroupsLoader.Load(atexit.Context())
}
