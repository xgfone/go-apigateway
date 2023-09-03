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

package loader

import (
	"context"
	"log/slog"
	"time"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
	"github.com/xgfone/go-apigateway/pkg/http/provider"
	"github.com/xgfone/go-apigateway/pkg/http/runtime"
)

// RouteProviderLoader returns a route loader based on the route provider.
func RouteProviderLoader(provider provider.RouteProvider, interval time.Duration) ResourceLoader {
	var lastetag string
	var lastroutes []dynamicconfig.Route

	return ResourceLoaderFunc(func(ctx context.Context, cb func(any)) {
		cb(lastroutes)
		load(ctx, interval, func() {
			routes, newtag, err := provider.Routes(lastetag)
			if err != nil {
				slog.Error("fail to get the routes from the route provider",
					"lastetag", lastetag, "err", err)
				return
			}

			if newtag == lastetag { // Not Changed
				return
			}

			adds, dels := dynamicconfig.DiffRoutes(routes, lastroutes)

			addRoutes := make([]runtime.Route, len(adds))
			for i, r := range adds {
				addRoutes[i], err = runtime.NewRoute(r)
				if err != nil {
					slog.Error("fail to build the route", "route", r, "err", err)
					return
				}
			}

			delRoutes := make([]runtime.Route, len(dels))
			for i, r := range dels {
				delRoutes[i] = runtime.Route{Route: r}
			}

			if len(addRoutes) > 0 {
				runtime.DefaultRouter.AddRoutes(addRoutes...)
				slog.Info("provider loader adds the routes", "routes", adds)
			}
			if len(delRoutes) > 0 {
				runtime.DefaultRouter.DelRoutes(delRoutes...)
				slog.Info("provider loader deletes the routes", "routes", dels)
			}

			lastroutes = routes
			lastetag = newtag
			cb(lastroutes)
		})
	})
}
