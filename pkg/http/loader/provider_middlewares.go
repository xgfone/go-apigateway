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
	"github.com/xgfone/go-apigateway/pkg/internal/maps"
)

// MiddlewareGroupProviderLoader returns a middleware group loader
// based on the middleware group provider.
func MiddlewareGroupProviderLoader(provider provider.MiddlewareGroupProvider, interval time.Duration) ResourceLoader {
	var lastetag string
	var lastgroups dynamicconfig.MiddlewareGroups

	return ResourceLoaderFunc(func(ctx context.Context, cb func(any)) {
		cb(lastgroups)
		load(ctx, interval, func() {
			groups, newtag, err := provider.MiddlewareGroups(lastetag)
			if err != nil {
				slog.Error("fail to get the middleware groups from the middleware group provider",
					"lastetag", lastetag, "err", err)
				return
			}

			if newtag == lastetag { // Not Changed
				return
			}

			changes := dynamicconfig.DiffMiddlewareGroups(groups, lastgroups)

			for gid, group := range changes {
				g := runtime.DefaultMiddlewareGroupManager.GetGroup(gid)
				if g == nil {
					mws, err := runtime.BuildMiddlewares(group.Adds)
					if err != nil {
						slog.Error("provider loader failed to build the group middlewares",
							"group", gid, "err", err)
					} else {
						g = runtime.NewMiddlewareGroup(mws...)
						runtime.DefaultMiddlewareGroupManager.AddGroup(gid, g)
						slog.Info("provider loader adds the new middleware group",
							"group", gid, "middlewares", group.Adds)
					}
					continue
				}

				if len(group.Dels) > 0 {
					g.DelMiddlewares(maps.Keys(group.Dels)...)
					slog.Info("provider loader deletes the group middlewares",
						"group", gid, "middlewares", group.Dels)
				}

				if len(group.Adds) > 0 {
					mws, err := runtime.BuildMiddlewares(group.Adds)
					if err != nil {
						slog.Error("provider loader failed to build the group middlewares",
							"group", gid, "err", err)
					} else {
						g.AddMiddlewares(mws...)
						slog.Info("provider loader adds the new group middlewares",
							"group", gid, "middlewares", group.Adds)
					}
				}
			}

			lastgroups = groups
			cb(lastgroups)
		})
	})
}
