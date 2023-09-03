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
	"fmt"
	"log/slog"
	"os"

	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-defaults"
)

// TODO: log file and configuration file

var loglevel = flag.String("log.level", "info", "The log level, such as debug, info, warn, error.")

func initlogging() {
	var level slog.Level
	if err := level.UnmarshalText([]byte(*loglevel)); err != nil {
		fatal("fail to parse the log level", "level", *loglevel, "err", err)
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		ReplaceAttr: replaceSourceAttr,
		AddSource:   true,
		Level:       level,
	})
	slog.SetDefault(slog.New(handler))
}

func replaceSourceAttr(groups []string, a slog.Attr) slog.Attr {
	switch {
	case a.Key == slog.SourceKey:
		if src, ok := a.Value.Any().(*slog.Source); ok {
			a.Value = slog.StringValue(fmt.Sprintf("%s:%d", defaults.TrimPkgFile(src.File), src.Line))
		}
	}
	return a
}

func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	atexit.Exit(1)
}
