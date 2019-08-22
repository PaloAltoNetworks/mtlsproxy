// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpproxy

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/aporeto-inc/mtlsproxy/internal/configuration"
)

func makeHandleHTTP(dest string) func(w http.ResponseWriter, req *http.Request) {

	u, err := url.Parse(dest)
	if err != nil {
		panic(err)
	}

	rewriteHost := u.Host
	rewriteSchema := u.Scheme

	return func(w http.ResponseWriter, req *http.Request) {

		req.URL.Host = rewriteHost
		req.URL.Scheme = rewriteSchema

		resp, err := http.DefaultTransport.RoundTrip(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		defer resp.Body.Close() // nolint: errcheck
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		w.WriteHeader(resp.StatusCode)

		if _, err = io.Copy(w, resp.Body); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}
}

// Start starts the proxy
func Start(cfg *configuration.Configuration, tlsConfig *tls.Config) {

	server := &http.Server{
		Addr:      cfg.ListenAddress,
		TLSConfig: tlsConfig,
		Handler:   http.HandlerFunc(makeHandleHTTP(cfg.Backend)),
	}

	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatalln("Unable to start proxy:", err)
		}
	}()

	log.Printf("MTLSProxy is ready. mode:%s listen:%s backend:%s ", cfg.Mode, cfg.ListenAddress, cfg.Backend)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
}
