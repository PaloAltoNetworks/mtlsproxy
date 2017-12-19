// Package configuration is a small package for handling configuration
package configuration

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/aporeto-inc/tg/tglib"
	"github.com/aporeto-inc/underwater/conf"

	"github.com/aporeto-inc/addedeffect/lombric"
	"github.com/aporeto-inc/mtlsproxy/internal/versions"
)

// Configuration hold the service configuration.
type Configuration struct {
	Backend                  string `mapstructure:"backend"         desc:"destination host"                                     default:"http://127.0.0.1"`
	ClientCAPoolPath         string `mapstructure:"clients-ca"      desc:"Path the CAs used to verify client certificates"      required:"true"`
	ListenAddress            string `mapstructure:"listen"          desc:"Listening address"                                    default:":443"`
	ServerCertificateKeyPass string `mapstructure:"cert-key-pass"   desc:"Password for the server certificate key"              `
	ServerCertificateKeyPath string `mapstructure:"cert-key"        desc:"Path to the server certificate key"                   required:"true"`
	ServerCertificatePath    string `mapstructure:"cert"            desc:"Path to the server certificate"                       required:"true"`

	conf.LoggingConf `mapstructure:",squash"`

	ClientCAPool      *x509.CertPool
	ServerCertificate tls.Certificate
}

// Prefix returns the configuration prefix.
func (c *Configuration) Prefix() string { return "mtlsproxy" }

// PrintVersion prints the current version.
func (c *Configuration) PrintVersion() {
	fmt.Printf("mtls - %s (%s)\n", versions.ProjectVersion, versions.ProjectSha)
}

// NewConfiguration returns a new configuration.
func NewConfiguration() *Configuration {

	c := &Configuration{}
	lombric.Initialize(c)

	data, err := ioutil.ReadFile(c.ClientCAPoolPath)
	if err != nil {
		panic(err)
	}
	c.ClientCAPool = x509.NewCertPool()
	c.ClientCAPool.AppendCertsFromPEM(data)

	keyData, err := ioutil.ReadFile(c.ServerCertificateKeyPath)
	if err != nil {
		panic(err)
	}
	keyBlock, err := tglib.DecryptPrivateKeyPEM(keyData, c.ServerCertificateKeyPass)
	if err != nil {
		panic(err)
	}

	certData, err := ioutil.ReadFile(c.ServerCertificatePath)
	if err != nil {
		panic(err)
	}
	c.ServerCertificate, err = tls.X509KeyPair(certData, pem.EncodeToMemory(keyBlock))
	if err != nil {
		panic(err)
	}

	return c
}
