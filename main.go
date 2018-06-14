package main

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/aporeto-inc/mtlsproxy/internal/configuration"
	"github.com/aporeto-inc/mtlsproxy/internal/versions"
	"go.aporeto.io/addedeffect/logutils"
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

func main() {

	cfg := configuration.NewConfiguration()
	logutils.Configure(cfg.LogLevel, cfg.LogFormat)

	banner(versions.ProjectVersion, versions.ProjectSha)
	time.Local = time.UTC

	tlsConfig := &tls.Config{
		ClientAuth:               tls.RequireAndVerifyClientCert,
		ClientCAs:                cfg.ClientCAPool,
		MinVersion:               tls.VersionTLS12,
		SessionTicketsDisabled:   true,
		PreferServerCipherSuites: true,
		Certificates:             []tls.Certificate{cfg.ServerCertificate},
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
		startHTTPProxy(cfg, tlsConfig)
	case "tcp":
		startTCPProxy(cfg, tlsConfig)
	}
}
