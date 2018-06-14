package httpproxy

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
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
			logrus.WithError(err).Fatal("Unable to start proxy")
		}
	}()

	logrus.
		WithField("mode", cfg.Mode).
		WithField("listen", cfg.ListenAddress).
		WithField("backend", cfg.Backend).
		Info("MTLSProxy is ready")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
}
