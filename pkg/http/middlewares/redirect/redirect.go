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

// Package redirect provides a redirect middleware.
package redirect

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-binder"
)

// Priority is the default priority of the middleware.
const Priority = 400

func init() {
	runtime.RegisterMiddlewareBuilder("redirect", func(name string, conf map[string]any) (runtime.Middleware, error) {
		var config Config
		if err := binder.BindStructToMap(&config, "json", conf); err != nil {
			return nil, err
		}
		return Redirect(name, Priority, &config)
	})
}

// Config is used to configure the redirect middleware.
type Config struct {
	// Optional, Default: 302
	Code int `json:"code" yaml:"code"`

	// Required, either-or

	HttpToHttps bool `json:"httpToHttps,omitempty" yaml:"httpToHttps,omitempty"`

	AppendQuery bool   `json:"appendQuery,omitempty" yaml:"appendQuery,omitempty"`
	Location    string `json:"location,omitempty" yaml:"location,omitempty"`
}

// Redirect returns a new middleware that returns the redirect response,
// such as 301/302/307/308, when the incoming request matches the given
// conditions.
func Redirect(name string, priority int, config *Config) (runtime.Middleware, error) {
	conf := *config
	if conf.Code == 0 {
		conf.Code = 302
	} else if conf.Code < 301 || conf.Code >= 400 {
		return nil, fmt.Errorf("Redirect: invalid redirect code %d", conf.Code)
	}

	var redirect func(*runtime.Context) bool
	switch {
	case config.HttpToHttps:
		redirect = httpsredirect{src: "mw:" + name, code: conf.Code}.Handle

	case config.Location != "":
		redirect = locredirect{
			src:      "mw:" + name,
			code:     conf.Code,
			location: config.Location,
			addquery: config.AppendQuery,
		}.Handle

	default:
		return nil, fmt.Errorf("Redirect: missing httpToHttps or location")
	}

	return runtime.NewMiddleware(name, priority, conf, func(next runtime.Handler) runtime.Handler {
		return func(c *runtime.Context) {
			if !redirect(c) {
				next(c)
			}
		}
	}), nil
}

type httpsredirect struct {
	src string

	code int
}

func (r httpsredirect) Handle(c *runtime.Context) (ok bool) {
	if !c.NeedModeForward(r.src, nil) {
		return
	}

	if ok = c.ClientRequest.TLS == nil; ok {
		host := strings.TrimPrefix(c.ClientRequest.Host, "http://")
		loc := strings.Join([]string{"https://", host, c.ClientRequest.RequestURI}, "")
		redirect(c.ClientResponse, r.code, loc)
	}
	return
}

type locredirect struct {
	src string

	code     int
	location string
	addquery bool
}

func (r locredirect) Handle(c *runtime.Context) (ok bool) {
	if !c.NeedModeForward(r.src, nil) {
		return
	}

	loc := r.location
	if r.addquery {
		loc = strings.Join([]string{loc, c.ClientRequest.URL.RawQuery}, "?")
	}
	redirect(c.ClientResponse, r.code, loc)
	return true
}

func redirect(w http.ResponseWriter, code int, location string) {
	w.Header().Set("Location", location)
	w.WriteHeader(code)
}
