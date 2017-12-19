package main

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/aporeto-inc/addedeffect/discovery"
	"github.com/aporeto-inc/addedeffect/logutils"
	"github.com/aporeto-inc/mtlsproxy/internal/configuration"
	"github.com/aporeto-inc/underwater/certification"
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

func main() {

	cfg := configuration.NewConfiguration()
	logutils.Configure(cfg.LogLevel, cfg.LogFormat)

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
		"mtlsproxy",
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
		Handler: http.HandlerFunc(makeHandleHTTP(cfg.ProxyDest)),
	}

	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			zap.L().Fatal("Unable to start proxy", zap.Error(err))
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
}
