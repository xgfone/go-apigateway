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

package discovery

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-checker"
	"github.com/xgfone/go-loadbalancer"
	"github.com/xgfone/go-loadbalancer/endpoint"
	"github.com/xgfone/go-loadbalancer/healthcheck"
)

// BuildStaticServer is used to customize the building of the static server.
//
// Default: use NewEndpoint to build it.
var BuildStaticServer func(server dynamicconfig.Server) (loadbalancer.Endpoint, error)

func buildUpstreamStaticServers(servers []dynamicconfig.Server) (loadbalancer.Endpoints, error) {
	endpoints := make(loadbalancer.Endpoints, len(servers))
	for i, s := range servers {
		ep, err := BuildStaticServer(s)
		if err != nil {
			return nil, err
		}
		endpoints[i] = ep
	}
	return endpoints, nil
}

// StaticDiscovery is the endpoint discovery based on the static servers.
type StaticDiscovery struct {
	name string

	// HealthCheck
	hcno bool // If true, disable healthcheck
	hccg healthcheck.Config
	reqc dynamicconfig.Request
	head http.Header
	tcp  bool

	// Delay to create them when starting the discovery.
	hc *healthcheck.HealthChecker
	em *endpoint.Manager
}

// NewStaticDiscovery returns a new endpoints discovery
// based on a set of static servers.
func NewStaticDiscovery(name string, sd dynamicconfig.StaticDiscovery) (*StaticDiscovery, error) {
	eps, err := buildUpstreamStaticServers(sd.Servers)
	if err != nil {
		return nil, err
	}
	_sd := &StaticDiscovery{name: name, em: endpoint.NewManager(len(eps))}
	_sd.em.Upserts(eps...)

	// Health Checker
	switch {
	case sd.HealthCheck == nil:
		_sd.hccg = healthcheck.DefaultConfig

	case sd.HealthCheck.Disable:
		_sd.hcno = true
		return _sd, nil

	default:
		if sd.HealthCheck.Request != nil {
			switch s := sd.HealthCheck.Request.Scheme; s {
			case "", "http", "https":
			case "tcp":
				_sd.tcp = true
			default:
				return nil, fmt.Errorf("unsupported healthcheck request scheme '%s'", s)
			}

			_sd.reqc = *sd.HealthCheck.Request
			_sd.reqc.Host = strings.ToUpper(_sd.reqc.Host)
			_sd.reqc.Method = strings.ToUpper(_sd.reqc.Method)
			if _len := len(_sd.reqc.Header); _len > 0 {
				_sd.head = make(http.Header, _len)
				for k, v := range _sd.reqc.Header {
					_sd.head[http.CanonicalHeaderKey(k)] = []string{v}
				}
			}
		}

		if sd.HealthCheck.Checker == nil {
			_sd.hccg = healthcheck.DefaultConfig
		} else {
			c := sd.HealthCheck.Checker
			_sd.hccg = healthcheck.Config{
				Failure:  c.Failure,
				Timeout:  c.Timeout * time.Second,
				Interval: c.Timeout * time.Second,
			}

			if _sd.hccg.Failure <= 0 {
				_sd.hccg.Failure = checker.DefaultConfig.Failure
			}
			if _sd.hccg.Timeout < 0 {
				_sd.hccg.Timeout = checker.DefaultConfig.Timeout
			}
			if _sd.hccg.Interval <= 0 {
				_sd.hccg.Interval = checker.DefaultConfig.Interval
			}
		}
	}

	return _sd, nil
}

// ------------------------------------------------------------------------- //
// Inspect the configration.

// Discover implements the interface endpoint.Discovery.
func (d *StaticDiscovery) Discover() *endpoint.Static { return d.em.Discover() }

// Len returns the length of all the endpoints.
func (d *StaticDiscovery) Len() int { return d.em.Len() }

// AllEndpoints returns all the endpoints.
func (d *StaticDiscovery) AllEndpoints() map[loadbalancer.Endpoint]bool { return d.em.All() }

// Range ranges all the endpoints until the range function returns false
// or all the endpoints are ranged.
func (d *StaticDiscovery) Range(f func(ep loadbalancer.Endpoint, online bool) bool) {
	d.em.Range(f)
}

// HealthCheck returns the healthcheck configuration.
func (d *StaticDiscovery) HealthCheck() any {
	if d.hcno {
		return nil
	}

	return map[string]any{
		"request": d.reqc,
		"checker": map[string]int{
			"failure":  d.hccg.Failure,
			"timeout":  int(d.hccg.Timeout / time.Second),
			"interval": int(d.hccg.Interval / time.Second),
		},
	}
}

// ------------------------------------------------------------------------- //

// Start starts the static discovery.
func (d *StaticDiscovery) Start() {
	if d.hcno {
		d.em.SetAllOnline(true)
	} else {
		d.hc = healthcheck.New(d.name)
		d.hc.OnChanged(d.onchanged)
		d.hc.SetConfig(d.hccg)
		d.hc.SetChecker(d.check)

		d.em.Range(func(ep loadbalancer.Endpoint, _ bool) bool {
			d.hc.AddTarget(ep.ID())
			return true
		})

		d.hc.Start()
	}
}

// Stop stops the static discovery.
func (d *StaticDiscovery) Stop() {
	if d.hc != nil {
		d.hc.Stop()
	}
}

// ------------------------------------------------------------------------- //
// HealthCheck

func (d *StaticDiscovery) onchanged(epid string, online bool) {
	slog.Info("update the endpoint server online", "epid", epid, "online", online)
	d.em.SetOnline(epid, online)
}

// Check is a proxy to check the health of the endpoint.
func (d *StaticDiscovery) check(ctx context.Context, epid string) (ok bool) {
	slog.Debug("start to check the endpoint", "discovery", d.name, "epid", epid)
	if d.tcp {
		return d.checkByTCP(ctx, epid)
	}
	return d.checkByHTTP(ctx, epid)
}

func (d *StaticDiscovery) checkByTCP(ctx context.Context, epid string) (ok bool) {
	ep, _ := d.em.Get(epid)
	if ep == nil {
		return false
	}

	if s, ok := ep.(interface {
		CheckTCP(c context.Context, https bool) bool
	}); ok {
		return s.CheckTCP(ctx, d.reqc.Scheme == "https")
	}

	defaultPort := "80"
	if d.reqc.Scheme == "https" {
		defaultPort = "443"
	}

	addr := epid
	if strings.IndexByte(addr, '.') > 0 { // Ipv4
		if strings.LastIndexByte(addr, ':') < 0 {
			addr = net.JoinHostPort(addr, defaultPort)
		}
	} else {
		if addr[0] != '[' {
			addr = net.JoinHostPort(addr, defaultPort)
		}
	}

	conn, err := net.Dial("tcp", addr)
	_ = conn.Close()
	ok = err == nil
	return
}

func (d *StaticDiscovery) checkByHTTP(ctx context.Context, epid string) (ok bool) {
	ep, _ := d.em.Get(epid)
	if ep == nil {
		return false
	}

	req := d.buildhttpreq()

	if s, ok := ep.(interface {
		CheckHTTP(context.Context, *http.Request) bool
	}); ok {
		return s.CheckHTTP(ctx, req)
	}

	req.Host = epid
	req.URL.Host = epid

	resp, err := runtime.DefaultHttpClient.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
	return err == nil && resp.StatusCode >= 200 && resp.StatusCode < 400
}

func (d *StaticDiscovery) buildhttpreq() *http.Request {
	req := &http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Scheme: "http", Path: "/"},
	}

	if d.reqc.Method != "" {
		req.Method = strings.ToUpper(d.reqc.Method)
	}
	if d.reqc.Scheme != "" {
		req.URL.Scheme = d.reqc.Scheme
	}
	if d.reqc.Host != "" {
		req.Host = d.reqc.Host
	}
	if d.reqc.Path != "" {
		req.URL.Path = d.reqc.Path
	}

	if len(d.head) > 0 {
		req.Header = maps.Clone(d.head)
	}

	return req
}
