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

package updater

import (
	"context"
	"log/slog"

	"github.com/xgfone/go-apigateway/http/router"
	"github.com/xgfone/go-apigateway/orch"
)

// SyncHttpRoutes receives the whole http route configurations,
// and synchronize them to the runtime.
func SyncHttpRoutes(ctx context.Context, config <-chan []orch.HttpRoute) {
	var lasts []orch.HttpRoute

	_sync(ctx, config, func(configs []orch.HttpRoute) {
		adds, dels := orch.DiffHttpRoutes(configs, lasts)

		addroutes := make([]router.Route, 0, len(adds))
		for _, c := range adds {
			route, err := c.Build()
			if err != nil {
				slog.Error("fail to build the http route", "route", c, "err", err)
				continue
			}

			addroutes = append(addroutes, route)
		}
		router.DefaultRouter.AddRoutes(addroutes...)

		delroutes := make([]string, len(dels))
		for i, c := range dels {
			delroutes[i] = c.Id
		}
		router.DefaultRouter.DelRoutesByIds(delroutes...)

		lasts = configs
	})
}
