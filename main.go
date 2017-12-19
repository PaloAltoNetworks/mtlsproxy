package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/aporeto-inc/addedeffect/discovery"
	"github.com/aporeto-inc/addedeffect/logutils"
	"github.com/aporeto-inc/apoctl/versions"
	"github.com/aporeto-inc/mtlsproxy/internal/configuration"
	"github.com/aporeto-inc/underwater/certification"
	"go.uber.org/zap"
)

func banner(version, revision string) {
	fmt.Printf(`
             _   _
   _ __ ___ | |_| |___ _ __  _ __ _____  ___   _
  | '_ . _ \| __| / __| '_ \| '__/ _ \ \/ / | | |
  | | | | | | |_| \__ \ |_) | | | (_) >  <| |_| |
  |_| |_| |_|\__|_|___/ .__/|_|  \___/_/\_\\__, |
                       |_|                  |___/

  MTLS Proxy Service (public)
  %s - %s
_______________________________________________________________

`, version, revision)
}

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

func main() {

	cfg := configuration.NewConfiguration()
	logutils.Configure(cfg.LogLevel, cfg.LogFormat)

	banner(versions.ProjectVersion, versions.ProjectSha)
	time.Local = time.UTC

	zap.L().Info("Discovering platform", zap.String("cid", cfg.CidURL))
	pf, err := discovery.DiscoverPlatform(cfg.CidURL, cfg.CidCACertPool, false)
	if err != nil {
		zap.L().Fatal("Unable to discover platform", zap.Error(err))
	}
	zap.L().Debug("Platform discovered", pf.Fields()...)

	rootCAPool, err := pf.RootCAPool()
	if err != nil {
		zap.L().Fatal("Unable to retrieve root CA pool", zap.Error(err))
	}

	systemCAPool, err := pf.SystemCAPool()
	if err != nil {
		zap.L().Fatal("Unable to retrieve system CA pool", zap.Error(err))
	}

	_, servicesCertKeyGenerator := certification.CreateServiceCertificates(
		cfg.BackendName,
		rootCAPool,
		pf,
		cfg.IssuingCertKeyPassword,
		false,
		true,
		cfg.DNSAltNames,
		cfg.IPAltNames,
		cfg.PublicCertKeyPassword,
	)

	server := &http.Server{
		Addr: cfg.ListenAddress,
		TLSConfig: &tls.Config{
			ClientAuth:               tls.RequireAndVerifyClientCert,
			ClientCAs:                systemCAPool,
			MinVersion:               tls.VersionTLS12,
			SessionTicketsDisabled:   true,
			PreferServerCipherSuites: true,
			GetCertificate:           servicesCertKeyGenerator,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
		Handler: http.HandlerFunc(makeHandleHTTP(cfg.Backend)),
	}

	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			zap.L().Fatal("Unable to start proxy", zap.Error(err))
		}
	}()

	zap.L().Info("MTLSProxy is fully ready", zap.String("listen", cfg.ListenAddress), zap.String("backend", cfg.Backend))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
}
