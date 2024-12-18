package cmd

import (
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/validate"
)

// serverConfig is the configuration for serving DNS requests.
type serverConfig struct {
	// ListenAddresses is the addresses server listens for requests.
	ListenAddresses []*ipPortConfig `yaml:"listen_addresses"`
}

// type check
var _ validate.Interface = (*serverConfig)(nil)

// Validate implements the [validate.Interface] interface for *serverConfig.
func (c *serverConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	return validate.Slice("listen_addresses", c.ListenAddresses)
}
