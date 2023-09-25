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

	"github.com/xgfone/go-apigateway/http/middleware"
	"github.com/xgfone/go-apigateway/http/router"
	"github.com/xgfone/go-apigateway/orch"
	"github.com/xgfone/go-apigateway/upstream"
	"github.com/xgfone/go-loadbalancer/endpoint"
)

var manageraddr = flag.String("manageraddr", "", "The address used by manager. If set, start it.")

func initmanager() {
	if addr := *manageraddr; addr != "" {
		regiseterReloadRoutes()
		regiseterLoaderRoutes()
		regiseterRuntimeRoutes()
		go startserver("gateway", addr, http.DefaultServeMux, false)
	}
}

func regiseterReloadRoutes() {
	http.HandleFunc("/apigateway/reload/certs", func(w http.ResponseWriter, r *http.Request) {
		reloadcert <- struct{}{}
	})
	http.HandleFunc("/apigateway/reload/upstreams", func(w http.ResponseWriter, r *http.Request) {
		reloadupstreams <- struct{}{}
	})
	http.HandleFunc("/apigateway/reload/http/routes", func(w http.ResponseWriter, r *http.Request) {
		reloadhttproutes <- struct{}{}
	})
	http.HandleFunc("/apigateway/reload/http/middlewares/groups", func(w http.ResponseWriter, r *http.Request) {
		reloadhttpmwgroups <- struct{}{}
	})
}

func regiseterLoaderRoutes() {
	http.HandleFunc("/apigateway/loader/upstreams", func(w http.ResponseWriter, r *http.Request) {
		sendjson(w, map[string]any{"upstreams": upstreamsloader.Resource()})
	})

	http.HandleFunc("/apigateway/loader/http/routes", func(w http.ResponseWriter, r *http.Request) {
		sendjson(w, map[string]any{"routes": httproutesloader.Resource()})
	})

	http.HandleFunc("/apigateway/loader/http/middlewares/groups", func(w http.ResponseWriter, r *http.Request) {
		sendjson(w, map[string]any{"groups": httpmwgroupsloader.Resource()})
	})
}

func sendjson(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(200)
	_ = json.NewEncoder(w).Encode(v)
}

func regiseterRuntimeRoutes() {
	http.HandleFunc("/apigateway/runtime/upstreams", func(w http.ResponseWriter, r *http.Request) {
		_ups := upstream.Manager.Gets()
		ups := make([]any, 0, len(_ups))
		for _, _up := range _ups {
			_eps := _up.Discover().Endpoints
			eps := make([]map[string]any, len(_eps))
			for i, ep := range _eps {
				eps[i] = map[string]any{
					"host":   ep.ID(),
					"weight": endpoint.GetWeight(ep),
				}
			}

			ups = append(ups, map[string]any{
				"id":        _up.Name(),
				"host":      _up.Host(),
				"scheme":    _up.Scheme(),
				"policy":    _up.GetBalancer().Policy(),
				"timeout":   _up.GetTimeout().String(),
				"endpoints": eps,
			})
		}
		sendjson(w, map[string][]any{"upstreams": ups})
	})

	http.HandleFunc("/apigateway/runtime/http/routes", func(w http.ResponseWriter, r *http.Request) {
		sendjson(w, map[string]any{"routes": router.DefaultRouter.Routes()})
	})

	http.HandleFunc("/apigateway/runtime/http/middlewares/groups", func(w http.ResponseWriter, r *http.Request) {
		_groups := middleware.DefaultGroupManager.Gets()
		groups := make(map[string]any, len(_groups))
		for id, _group := range _groups {
			_mws := _group.Middlewares()
			mws := make(map[string]any, len(_mws))
			for _, _mw := range _mws {
				mws[_mw.Name()] = orch.GetConfig(_mw)
			}
			groups[id] = mws
		}
		sendjson(w, map[string]any{"groups": groups})
	})
}
