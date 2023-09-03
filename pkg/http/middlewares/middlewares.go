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

// Package middlewares registers the inner middlewares.
package middlewares

import (
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares/allow"
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares/auth"
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares/auth/forward"
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares/block"
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares/circuitbreaker"
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares/cors"
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares/gzip"
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares/processor"
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares/ratelimit"
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares/redirect"
	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares/transformbody"
)
