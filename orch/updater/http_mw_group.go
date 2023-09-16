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

	"github.com/xgfone/go-apigateway/http/middleware"
	"github.com/xgfone/go-apigateway/orch"
)

// SyncHttpMiddlewareGroups receives the whole http middleware group configurations,
// and synchronize them to the runtime.
func SyncHttpMiddlewareGroups(ctx context.Context, config <-chan []orch.MiddlewareGroup) {
	var lasts []orch.MiddlewareGroup
	_sync(ctx, config, func(configs []orch.MiddlewareGroup) {
		adds, dels := orch.DiffMiddlewareGroups(configs, lasts)

		addgoups := make(map[string]*middleware.Group, len(adds))
		for _, c := range adds {
			group, err := c.HttpBuild()
			if err != nil {
				slog.Error("fail to build the http middleware group", "group", c, "err", err)
				continue
			}

			addgoups[group.Name()] = group
			slog.Info("build the http middleware group and later add or update it", "group", c)
		}
		middleware.DefaultGroupManager.Adds(addgoups)

		delgroups := make([]string, len(dels))
		for i, group := range dels {
			delgroups[i] = group.Name
			slog.Info("later delete the http middleware group", "group", group.Name)
		}
		middleware.DefaultGroupManager.Dels(delgroups...)

		lasts = configs
	})
}
