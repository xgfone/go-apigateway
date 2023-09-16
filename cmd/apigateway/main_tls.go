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
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xgfone/go-apigateway/logger"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-tlsx"
)

var (
	tlsjsonfile = flag.String("tls.jsonfile", "", "If set, add all the certificates in the file to the server.")
	tlscertfile = flag.String("tls.certfile", "", "The path of the certificate file. If set, start api gateway with HTTPS.")
	tlskeyfile  = flag.String("tls.keyfile", "", "The path of the certificate key file. If set, start api gateway with HTTPS.")

	reloadcert = make(chan struct{}, 1)
)

func tryTLSListener(ln net.Listener) net.Listener {
	if tlsconfig := getTLSConfig(); tlsconfig != nil {
		ln = tls.NewListener(ln, tlsconfig)
	} else {
		go loopreloadcert()
	}
	return ln
}

func loopreloadcert() {
	for {
		select {
		case <-reloadcert:
		case <-atexit.Done():
			return
		}
	}
}

func getTLSConfig() *tls.Config {
	files := getTLSCertFiles(*tlscertfile, *tlskeyfile, *tlsjsonfile)
	if len(files) == 0 {
		return nil
	}

	cb := func(name string) func(*tls.Certificate) {
		return func(c *tls.Certificate) { tlsx.DefaultCertManager.Add(name, c) }
	}

	for i, file := range files {
		name := file.Name
		if name == "" {
			key := strings.TrimSuffix(file.KeyFile, filepath.Ext(file.KeyFile))
			cert := strings.TrimSuffix(file.CertFile, filepath.Ext(file.CertFile))
			if key == cert {
				name = cert
			} else {
				name = fmt.Sprintf("cert%d", i+1)
			}
		}
		go tlsx.WatchCertFiles(atexit.Context(), reloadcert, time.Minute, file.CertFile, file.KeyFile, cb(name))
	}

	return tlsx.DefaultCertManager.ServerConfig(nil)
}

type keycertfile struct {
	Name     string `json:"name"`
	KeyFile  string `json:"keyFile"`
	CertFile string `json:"certFile"`
}

func getTLSCertFiles(certfile, keyfile, jsonfile string) (files []keycertfile) {
	if certfile != "" && keyfile != "" {
		files = append(files, keycertfile{CertFile: certfile, KeyFile: keyfile})
	}

	if jsonfile != "" {
		data, err := os.ReadFile(jsonfile)
		if err != nil {
			logger.Fatal("fail to read the tls json config file", "file", jsonfile, "err", err)
		}

		if len(data) > 0 {
			if len(data) > 0 {
				var results []keycertfile
				if err := json.Unmarshal(data, &results); err != nil {
					logger.Fatal("fail to parse the tls json config file", "file", jsonfile, "err", err)
				}
				files = append(files, results...)
			}
		}
	}

	return
}
