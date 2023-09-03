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
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
)

// TODO:
// tls.Config:
//     ServerName(string): Client: vhost
//     InsecureSkipVerify(bool)
//     RootCAs: Client verifies the certificate of Server
//     Certs:

var (
	tlsjsonfile = flag.String("tls.jsonfile", "", "If set, add all the certificates in the file to the server.")
	tlscertfile = flag.String("tls.certfile", "", "The path of the certificate file.")
	tlskeyfile  = flag.String("tls.keyfile", "", "The path of the certificate key file.")
)

func tryTLSListener(ln net.Listener) net.Listener {
	if tlsconfig := getTLSConfig(); tlsconfig != nil {
		ln = tls.NewListener(ln, tlsconfig)
	}
	return ln
}

func getTLSConfig() (tlsconfig *tls.Config) {
	files := getTLSCertFiles(*tlscertfile, *tlskeyfile, *tlsjsonfile)
	if len(files) > 0 {
		certs, err := parseCertificateFiles(files...)
		if err != nil {
			fatal("fail to parse the certificate files", "files", files, "err", err)
		}

		return &tls.Config{
			GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
				for i, _len := 0, len(certs); i < _len; i++ {
					if err := chi.SupportsCertificate(certs[i]); err == nil {
						return certs[i], nil
					}
				}
				return nil, fmt.Errorf("tls: no proper certificate is configured for '%s'", chi.ServerName)
			},
		}
	}
	return
}

func parseCertificate(certpem, keypem []byte) (tls.Certificate, error) {
	tlscert, err := tls.X509KeyPair(certpem, keypem)
	if err != nil {
		return tlscert, err
	}

	if tlscert.Leaf == nil {
		tlscert.Leaf, err = x509.ParseCertificate(tlscert.Certificate[0])
		if err != nil {
			return tlscert, err
		}
	}

	return tlscert, nil
}

func parseCertificateFiles(files ...keycertfile) ([]*tls.Certificate, error) {
	certs := make([]*tls.Certificate, len(files))
	for i, f := range files {
		certpem, err := os.ReadFile(f.CertFile)
		if err != nil {
			return nil, err
		}

		keypem, err := os.ReadFile(f.KeyFile)
		if err != nil {
			return nil, err
		}

		cert, err := parseCertificate(certpem, keypem)
		if err != nil {
			return nil, err
		}

		certs[i] = &cert
	}
	return certs, nil
}

type keycertfile struct {
	CertFile string
	KeyFile  string
}

func getTLSCertFiles(certfile, keyfile, jsonfile string) (files []keycertfile) {
	if certfile != "" && keyfile != "" {
		files = append(files, keycertfile{CertFile: certfile, KeyFile: keyfile})
	}

	if jsonfile != "" {
		data, err := os.ReadFile(jsonfile)
		if err != nil {
			fatal("fail to read the tls json config file", "file", jsonfile, "err", err)
		}

		if len(data) > 0 {
			// data = bytex.RemoveComments(data)
			if len(data) > 0 {
				var results []keycertfile
				if err := json.Unmarshal(data, &results); err != nil {
					fatal("fail to parse the tls json config file", "file", jsonfile, "err", err)
				}
				files = append(files, results...)
			}
		}
	}

	return
}
