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

	"github.com/xgfone/go-apigateway/orch"
	"github.com/xgfone/go-apigateway/upstream"
)

// SyncUpstreams receives the whole upstream configurations,
// and synchronize them to the runtime.
func SyncUpstreams(ctx context.Context, config <-chan []orch.Upstream) {
	var lasts []orch.Upstream
	_sync(ctx, config, func(configs []orch.Upstream) {
		adds, dels := orch.DiffUpstreams(configs, lasts)

		addups := make(map[string]*upstream.Upstream, len(adds))
		for _, c := range adds {
			up, err := c.Build()
			if err != nil {
				slog.Error("fail to build the upstream", "upstream", c, "err", err)
				continue
			}

			addups[up.Name()] = up
			slog.Info("build upstream and later add or update it", "upstream", c)
		}
		upstream.Manager.Adds(addups)

		delups := make([]string, len(dels))
		for i, c := range dels {
			delups[i] = c.Id
			slog.Info("later delete the upstream", "upstreamid", c.Id)
		}
		upstream.Manager.Dels(delups...)

		lasts = configs
	})
}
