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
	"encoding/json"
	"flag"
	"net/http"
	"time"

	"github.com/xgfone/go-apigateway/pkg/http/discovery"
	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-loadbalancer"
)

var manageraddr = flag.String("manageraddr", "", "The address used by manager.")

func initmanager() {
	if *manageraddr != "" {
		registerProviderRoutes()
		regiseterRuntimeRoutes()
		go startserver(*manageraddr, http.DefaultServeMux, false)
	}
}

func sendjson(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(200)
	_ = json.NewEncoder(w).Encode(v)
}

func registerProviderRoutes() {
	http.HandleFunc("/apigateway/provider/routes", func(w http.ResponseWriter, r *http.Request) {
		sendjson(w, routesLoader.Resource())
	})

	http.HandleFunc("/apigateway/provider/upstreams", func(w http.ResponseWriter, r *http.Request) {
		sendjson(w, upstreamsLoader.Resource())
	})

	http.HandleFunc("/apigateway/provider/middlewares/groups", func(w http.ResponseWriter, r *http.Request) {
		sendjson(w, mdwgroupsLoader.Resource())
	})
}

func regiseterRuntimeRoutes() {
	http.HandleFunc("/apigateway/runtime/routes", func(w http.ResponseWriter, r *http.Request) {
		sendjson(w, runtime.DefaultRouter.Routes())
	})

	http.HandleFunc("/apigateway/runtime/upstreams", func(w http.ResponseWriter, r *http.Request) {
		_ups := runtime.GetUpstreams()
		ups := make([]any, len(_ups))
		for i, _up := range _ups {
			f := _up.Forwarder()
			d := f.GetDiscovery().(*discovery.StaticDiscovery)
			eps := make([]any, 0, d.Len())
			d.Range(func(ep loadbalancer.Endpoint, online bool) bool {
				var config any
				if v, ok := ep.(interface{ Config() any }); ok {
					config = v.Config()
				}
				eps = append(eps, map[string]any{
					"id":     ep.ID(),
					"config": config,
					"online": online,
				})
				return true
			})

			ups[i] = map[string]any{
				"id":          _up.UpConfig.Id,
				"policy":      f.GetBalancer().Policy(),
				"timeout":     f.GetTimeout() / time.Second,
				"endpoints":   eps,
				"healthCheck": d.HealthCheck(),
			}
		}
		sendjson(w, ups)
	})

	http.HandleFunc("/apigateway/runtime/middlewares/groups", func(w http.ResponseWriter, r *http.Request) {
		_groups := runtime.DefaultMiddlewareGroupManager.Groups()
		groups := make(map[string]any, len(_groups))
		for id, _group := range _groups {
			_mws := _group.Middlewares()
			mws := make(map[string]any, len(_mws))
			for _, _mw := range _mws {
				mws[_mw.Name()] = runtime.GetConfig(_mw)
			}
			groups[id] = mws
		}
		sendjson(w, groups)
	})
}
