package main

import (
	"crypto/tls"
	"time"

	"github.com/aporeto-inc/mtlsproxy/internal/configuration"
	"github.com/aporeto-inc/mtlsproxy/internal/httpproxy"
	"github.com/aporeto-inc/mtlsproxy/internal/tcpproxy"
)

func main() {

	cfg := configuration.NewConfiguration()

	time.Local = time.UTC

	tlsConfig := &tls.Config{
		ClientAuth:               tls.RequireAndVerifyClientCert,
		ClientCAs:                cfg.ClientCAPool,
		MinVersion:               tls.VersionTLS12,
		SessionTicketsDisabled:   true,
		PreferServerCipherSuites: true,
		Certificates:             cfg.ServerCertificates,

		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	switch cfg.Mode {
	case "http":
		httpproxy.Start(cfg, tlsConfig)
	case "tcp":
		tcpproxy.Start(cfg, tlsConfig)
	}
}
