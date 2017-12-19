// Package configuration is a small package for handling configuration
package configuration

import (
	"fmt"

	"github.com/aporeto-inc/addedeffect/lombric"
	"github.com/aporeto-inc/mtlsproxy/internal/versions"
	"github.com/aporeto-inc/underwater/conf"
)

// Configuration hold the service configuration.
type Configuration struct {
	ProxyDest   string `mapstructure:"dest"             desc:"destination host"                                     default:"http://127.0.0.1"`
	BackendName string `mapstructure:"backend-name"     desc:"name of the backend that will be used in certificate" default:"mtlsproxy"`

	conf.APIServerConf  `mapstructure:",squash"`
	conf.CidConf        `mapstructure:",squash"`
	conf.LoggingConf    `mapstructure:",squash"`
	conf.TLSIssuingConf `mapstructure:",squash" override:"dns-alt-name=mtlsproxy"`
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

	return c
}
