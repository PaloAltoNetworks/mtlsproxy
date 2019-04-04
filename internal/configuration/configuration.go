// Package configuration is a small package for handling configuration
package configuration

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"go.aporeto.io/addedeffect/lombric"
	"go.aporeto.io/tg/tglib"
)

// Configuration hold the service configuration.
type Configuration struct {
	Backend                  string `mapstructure:"backend"         desc:"destination host"                                     default:"http://127.0.0.1"`
	ClientCAPoolPath         string `mapstructure:"clients-ca"      desc:"Path the CAs used to verify client certificates"      required:"true"`
	ListenAddress            string `mapstructure:"listen"          desc:"Listening address"                                    default:":443"`
	ServerCertificateKeyPass string `mapstructure:"cert-key-pass"   desc:"Password for the server certificate key"              `
	ServerCertificateKeyPath string `mapstructure:"cert-key"        desc:"Path to the server certificate key"                   required:"true"`
	ServerCertificatePath    string `mapstructure:"cert"            desc:"Path to the server certificate"                       required:"true"`
	Mode                     string `mapstructure:"mode"            desc:"Proxy mode"                                           default:"http" allowed:"tcp,http"`
	LogFormat                string `mapstructure:"log-format"      desc:"Log format"                                           default:"console"`
	LogLevel                 string `mapstructure:"log-level"       desc:"Log level"                                            default:"info"`

	ClientCAPool       *x509.CertPool
	ServerCertificates []tls.Certificate
}

// Prefix returns the configuration prefix.
func (c *Configuration) Prefix() string { return "mtlsproxy" }

// PrintVersion prints the current version.
func (c *Configuration) PrintVersion() {
	fmt.Printf("mtls - %s (%s)\n", "1.0", "")
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

	certs, key, err := tglib.ReadCertificatePEMs(c.ServerCertificatePath, c.ServerCertificateKeyPath, c.ServerCertificateKeyPass)
	if err != nil {
		panic(err)
	}

	tc, err := tglib.ToTLSCertificates(certs, key)
	if err != nil {
		panic(err)
	}
	c.ServerCertificates = append(c.ServerCertificates, tc)

	return c
}
