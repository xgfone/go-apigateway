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

// UpstreamProviderLoader returns an upstream loader based on the upstream provider.
func UpstreamProviderLoader(provider provider.UpstreamProvider, interval time.Duration) ResourceLoader {
	var lastetag string
	var lastups []dynamicconfig.Upstream

	return ResourceLoaderFunc(func(ctx context.Context, cb func(any)) {
		cb(lastups)
		load(ctx, interval, func() {
			ups, newtag, err := provider.Upstreams(lastetag)
			if err != nil {
				slog.Error("fail to get the upstreams from the upstream provider",
					"lastetag", lastetag, "err", err)
				return
			}

			if newtag == lastetag { // Not Changed
				return
			}

			adds, dels := dynamicconfig.DiffUpstreams(ups, lastups)

			addUpstreams := make([]*runtime.Upstream, len(adds))
			for i, up := range adds {
				addUpstreams[i], err = runtime.NewUpstream(up)
				if err != nil {
					slog.Error("fail to build the upstream", "upstream", up, "err", err)
					return
				}
			}

			if len(dels) > 0 {
				for _, up := range dels {
					runtime.DelUpstream(up.Id)
				}
				slog.Info("provider loader deletes the upstreams", "upstreams", dels)
			}

			if len(addUpstreams) > 0 {
				for _, up := range addUpstreams {
					runtime.AddUpstream(up)
				}
				slog.Info("provider loader adds the upstreams", "upstreams", adds)
			}

			lastups = ups
			lastetag = newtag
			cb(lastups)
		})
	})
}
