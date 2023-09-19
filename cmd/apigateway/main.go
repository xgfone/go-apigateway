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
	"net"
	"net/http"

	"github.com/xgfone/go-apigateway/http/router"
	"github.com/xgfone/go-apigateway/http/server"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-atexit/signal"
	"github.com/xgfone/go-defaults"
)

var gatewayaddr = flag.String("gatewayaddr", ":80", "The address used by gateway.")

func init() { defaults.ExitFunc.Set(atexit.Exit) }

func main() {
	flag.Parse()
	initlogging()
	initmanager()
	initloader()
	go signal.WaitExit(atexit.Execute)
	startserver(*gatewayaddr, router.DefaultRouter, true)
	atexit.Wait() // wait that all the clean functions end.
}

func startserver(addr string, handler http.Handler, trytls bool) {
	svr := server.New(addr, handler)
	svr.RegisterOnShutdown(atexit.Execute)
	atexit.OnExit(func() { server.Stop(svr) })

	var cb func(net.Listener) net.Listener
	if trytls {
		cb = tryTLSListener
	}
	server.StartWithListenerCallback(svr, cb)
}
