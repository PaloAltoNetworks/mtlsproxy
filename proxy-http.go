package main

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/aporeto-inc/mtlsproxy/internal/configuration"
	"go.uber.org/zap"
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

func startHTTPProxy(cfg *configuration.Configuration, tlsConfig *tls.Config) {

	server := &http.Server{
		Addr:      cfg.ListenAddress,
		TLSConfig: tlsConfig,
		Handler:   http.HandlerFunc(makeHandleHTTP(cfg.Backend)),
	}

	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			zap.L().Fatal("Unable to start proxy", zap.Error(err))
		}
	}()

	zap.L().Info("MTLSProxy is ready",
		zap.String("mode", cfg.Mode),
		zap.String("listen", cfg.ListenAddress),
		zap.String("backend", cfg.Backend),
	)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
}
