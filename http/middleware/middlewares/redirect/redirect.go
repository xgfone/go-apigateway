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

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
)

func init() {
	middleware.DefaultRegistry.Register("redirect", func(name string, conf any) (middleware.Middleware, error) {
		var config Config
		if err := middleware.BindConf(name, &config, conf); err != nil {
			return nil, err
		}
		return Redirect(config)
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

// Redirect returns a new middleware named "redirect"
// that returns the redirect response, such as 301/302/307/308,
// when the incoming request matches the given conditions.
func Redirect(config Config) (middleware.Middleware, error) {
	if config.Code == 0 {
		config.Code = 302
	} else if config.Code < 301 || config.Code >= 400 {
		return nil, fmt.Errorf("Redirect: invalid redirect code %d", config.Code)
	}

	var redirect func(*core.Context) bool
	switch {
	case config.HttpToHttps:
		redirect = httpsredirect{code: config.Code}.Handle

	case config.Location != "":
		redirect = locredirect{
			code:     config.Code,
			location: config.Location,
			addquery: config.AppendQuery,
		}.Handle

	default:
		return nil, fmt.Errorf("Middleware<redirect>: missing httpToHttps or location")
	}

	return middleware.New("redirect", config, func(next core.Handler) core.Handler {
		return func(c *core.Context) {
			if c.IsAborted {
				return
			}

			if !redirect(c) {
				next(c)
			}
		}
	}), nil
}

type httpsredirect struct {
	code int
}

func (r httpsredirect) Handle(c *core.Context) (ok bool) {
	if ok = c.ClientRequest.TLS == nil; ok {
		host := strings.TrimPrefix(c.ClientRequest.Host, "http://")
		loc := strings.Join([]string{"https://", host, c.ClientRequest.RequestURI}, "")
		redirect(c.ClientResponse, r.code, loc)
	}
	return
}

type locredirect struct {
	location string
	addquery bool
	code     int
}

func (r locredirect) Handle(c *core.Context) (ok bool) {
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
