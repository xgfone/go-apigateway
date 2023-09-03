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
	"context"
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-apigateway/pkg/nets"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-atexit/signal"
)

var gatewayaddr = flag.String("gatewayaddr", ":80", "The address used by gateway.")

func main() {
	flag.Parse()
	initlogging()
	initloader()
	initmanager()
	go signal.WaitExit(atexit.Execute)
	startserver(*gatewayaddr, runtime.DefaultRouter, true)
}

func startserver(addr string, handler http.Handler, trytls bool) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fatal("fail to listen on the address", "protocol", "tcp", "addr", addr, "err", err)
		return
	}

	if trytls {
		ln = tryTLSListener(ln)
	}

	server := newServer(handler)
	server.RegisterOnShutdown(atexit.Execute)
	atexit.OnExit(func() { _ = server.Shutdown(context.Background()) })

	slog.Info("start the http server", "addr", addr)
	_ = server.Serve(ln)
	slog.Info("stop the http server", "addr", addr)
}

func newServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler: handler,

		ReadTimeout:  0,
		WriteTimeout: 0,

		IdleTimeout:       time.Minute * 3,
		ReadHeaderTimeout: time.Second * 3,

		ErrorLog:    errLogger(),
		BaseContext: baseContext,
		ConnContext: connContext,
	}
}

func errLogger() *log.Logger {
	return slog.NewLogLogger(slog.Default().Handler(), slog.LevelError)
}

func baseContext(ln net.Listener) context.Context {
	return nets.SetListenerIntoContext(atexit.Context(), ln)
}

func connContext(ctx context.Context, conn net.Conn) context.Context {
	return nets.SetConnIntoContext(ctx, conn)
}
