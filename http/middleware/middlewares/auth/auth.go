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

// Package auth provides a common auth middleware
// based on the request header 'Authorization'.
package auth

import (
	"errors"
	"maps"
	"net/http"
	"strings"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
	"github.com/xgfone/go-apigateway/http/statuscode"
)

var (
	errMissingAuthorization = errors.New("missing header 'Authorization'")
	errMissingAuthType      = errors.New("Authorization: missing auth type")
	errInvalidAuthType      = errors.New("Authorization: invalid auth type")
	errInvalidAuth          = errors.New("Authorization: invalid auth")
)

func init() {
	middleware.DefaultRegistry.Register("auth", func(name string, conf any) (middleware.Middleware, error) {
		var config struct {
			Type  string                       `json:"type" yaml:"type"`
			Auths map[string]map[string]string `json:"auths" yaml:"auths"`
		}
		if err := middleware.BindConf(name, &config, conf); err != nil {
			return nil, err
		}

		auth := authconfig{Type: config.Type, Auths: make(map[string]http.Header, len(config.Auths))}
		for value, infos := range config.Auths {
			header := make(http.Header, len(infos))
			for key, value := range infos {
				header.Set(key, value)
			}
			auth.Auths[value] = header
		}
		return Authorization(auth.handle), nil
	})
}

type authconfig struct {
	Type  string
	Auths map[string]http.Header
}

func (c authconfig) handle(_type, value string, header http.Header) (err error) {
	switch {
	case _type == "":
		err = errMissingAuthType

	case _type != c.Type:
		err = errInvalidAuthType

	default:
		if infos, ok := c.Auths[value]; !ok {
			err = errInvalidAuth
		} else if len(infos) > 0 {
			maps.Copy(header, infos)
		}
	}

	return
}

// Authorization returns a common middleware named "auth",
// which parses the request header "Authorization" to two fields,
// authtype(optional) and authvalue(required), and calls the handle function
// to finish the authorization that it maybe put them into the header
// if there some auth information to be passed to the upstream server.
func Authorization(handle func(authtype, authvalue string, header http.Header) error) middleware.Middleware {
	if handle == nil {
		panic("Authorization: the handle function must not be nil")
	}

	return middleware.New("auth", nil, func(next core.Handler) core.Handler {
		return func(c *core.Context) {
			if c.IsAborted {
				return
			}

			value := strings.TrimSpace(c.ClientRequest.Header.Get("Authorization"))
			if value == "" {
				c.Abort(statuscode.ErrUnauthorized.WithError(errMissingAuthorization))
				return
			}

			var _type string
			if index := strings.IndexByte(value, ' '); index > -1 {
				_type = value[:index]
				value = strings.TrimSpace(value[index+1:])
			}

			if err := handle(_type, value, c.ClientRequest.Header); err != nil {
				c.Abort(statuscode.ErrUnauthorized.WithError(err))
			} else {
				next(c)
			}
		}
	})
}
