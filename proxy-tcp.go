package main

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/Sirupsen/logrus"
	"go.uber.org/zap"

	"github.com/aporeto-inc/mtlsproxy/internal/configuration"
)

type proxy struct {
	from      string
	to        string
	tlsConfig *tls.Config
}

func newProxy(from, to string, tlsConfig *tls.Config) *proxy {

	return &proxy{
		from:      from,
		to:        to,
		tlsConfig: tlsConfig,
	}
}

func (p *proxy) start(ctx context.Context) error {

	listener, err := tls.Listen("tcp", p.from, p.tlsConfig)
	if err != nil {
		return err
	}

	for {
		select {

		default:
			if connection, err := listener.Accept(); err == nil {
				go p.handle(ctx, connection)
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (p *proxy) handle(ctx context.Context, connection net.Conn) {

	defer connection.Close() // nolint
	remote, err := net.Dial("tcp", p.to)
	if err != nil {
		return
	}
	defer remote.Close() // nolint

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go p.copy(ctx, remote, connection, wg)
	go p.copy(ctx, connection, remote, wg)

	wg.Wait()
}

func (p *proxy) copy(ctx context.Context, from, to net.Conn, wg *sync.WaitGroup) {

	defer wg.Done()

	select {

	default:
		if _, err := io.Copy(to, from); err != nil {
			logrus.WithError(err).Error("Error during data copy")
			return
		}

	case <-ctx.Done():
		return
	}
}

func startTCPProxy(cfg *configuration.Configuration, tlsConfig *tls.Config) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := newProxy(cfg.ListenAddress, cfg.Backend, tlsConfig).start(ctx); err != nil {
			logrus.WithError(err).Fatal("Unable to start proxy")
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
	cancel()
}
