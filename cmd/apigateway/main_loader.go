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

	_ "github.com/xgfone/go-apigateway/http/middleware/middlewares"

	"github.com/xgfone/go-apigateway/loader/dirloader"
	"github.com/xgfone/go-apigateway/orch"
	"github.com/xgfone/go-apigateway/orch/updater"
	"github.com/xgfone/go-atexit"
)

var (
	_                    = flag.String("provider", "localfiledir", "The provider of the dynamic configurations.")
	upstreamslocaldir    = flag.String("provider.localfiledir.upstreams", "upstreams", "The directory of the local files storing the upstreams.")
	httprouteslocaldir   = flag.String("provider.localfiledir.httproutes", "httproutes", "The directory of the local files storing the http routes.")
	httpmwgroupslocaldir = flag.String("provider.localfiledir.httpmiddlewaregroups", "httpmiddlewaregroups", "The directory of the local files storing the http middleware groups.")
	localfiledirinterval = flag.Duration("provider.localfiledir.interval", time.Minute, "The interval duration to check and reload the configurations.")
)

var (
	reloadconf         = make(chan struct{}, 1)
	upstreamsloader    *dirloader.DirLoader[orch.Upstream]
	httproutesloader   *dirloader.DirLoader[orch.HttpRoute]
	httpmwgroupsloader *dirloader.DirLoader[orch.MiddlewareGroup]
)

func initloader() {
	interval := *localfiledirinterval
	upstreamsch := make(chan []orch.Upstream)
	upstreamsloader = dirloader.New[orch.Upstream](*upstreamslocaldir)
	go updater.SyncUpstreams(atexit.Context(), upstreamsch)
	go upstreamsloader.Sync(atexit.Context(), "upstreams", interval, reloadconf, func(ups []orch.Upstream) (changed bool) {
		changed = updateUpstreams(ups)
		upstreamsch <- ups
		return
	})

	httproutesch := make(chan []orch.HttpRoute)
	httproutesloader = dirloader.New[orch.HttpRoute](*httprouteslocaldir)
	go updater.SyncHttpRoutes(atexit.Context(), httproutesch)
	go httproutesloader.Sync(atexit.Context(), "httproutes", interval, reloadconf, func(routes []orch.HttpRoute) (changed bool) {
		changed = updateHttpRoutes(routes)
		httproutesch <- routes
		return
	})

	httpmwgroupsch := make(chan []orch.MiddlewareGroup)
	httpmwgroupsloader = dirloader.New[orch.MiddlewareGroup](*httpmwgroupslocaldir)
	go updater.SyncHttpMiddlewareGroups(atexit.Context(), httpmwgroupsch)
	go httpmwgroupsloader.Sync(atexit.Context(), "httpmiddlewaregroups", interval, reloadconf, func(groups []orch.MiddlewareGroup) bool {
		httpmwgroupsch <- groups
		return false
	})
}

func updateUpstreams(ups []orch.Upstream) (changed bool) {
	for i, _len := 0, len(ups); i < _len; i++ {
		up := &ups[i]

		if up.Timeout > 0 && up.Timeout < time.Second {
			up.Timeout *= time.Millisecond
			changed = true
		}

		if up.Retry.Interval > 0 && up.Retry.Interval < time.Second {
			up.Retry.Interval *= time.Millisecond
			changed = true
		}
	}
	return
}

func updateHttpRoutes(routes []orch.HttpRoute) (changed bool) {
	for i, _len := 0, len(routes); i < _len; i++ {
		r := &routes[i]

		if r.RequestTimeout > 0 && r.RequestTimeout < time.Second {
			r.RequestTimeout *= time.Millisecond
			changed = true
		}

		if r.ForwardTimeout > 0 && r.ForwardTimeout < time.Second {
			r.ForwardTimeout *= time.Millisecond
			changed = true
		}
	}
	return
}
