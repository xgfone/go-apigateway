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
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/xgfone/go-apigateway/nets"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-atexit/signal"
	"github.com/xgfone/go-defaults"
)

func init() {
	defaults.ExitFunc.Set(atexit.Exit)
	go signal.WaitExit(atexit.Execute)
}

func startserver(name, addr string, handler http.Handler, trytls bool) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("fail to open the listener on the address",
			"protocol", "tcp", "name", name, "addr", addr, "err", err)
		return
	}

	if trytls {
		ln = tryTLSListener(ln)
	}

	svr := &http.Server{
		Addr:    addr,
		Handler: handler,

		IdleTimeout:       time.Minute * 3,
		ReadHeaderTimeout: time.Second * 3,

		ErrorLog:    slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
		BaseContext: baseContext,
		ConnContext: connContext,
	}

	atexit.OnExit(func() { _ = svr.Shutdown(context.Background()) })
	slog.Info("start the http server", "name", name, "addr", addr)
	defer slog.Info("stop the http server", "name", name, "addr", addr)

	_ = svr.Serve(ln)
	atexit.Wait() // Wait until all the clean functions end.
}

func baseContext(ln net.Listener) context.Context {
	return nets.SetListenerIntoContext(context.Background(), ln)
}

func connContext(ctx context.Context, conn net.Conn) context.Context {
	return nets.SetConnIntoContext(ctx, conn)
}
