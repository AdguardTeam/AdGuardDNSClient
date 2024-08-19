package cmd

import (
	"fmt"

	"github.com/AdguardTeam/golibs/errors"
)

// serverConfig is the configuration for serving DNS requests.
type serverConfig struct {
	// ListenAddresses is the addresses server listens for requests.
	ListenAddresses ipPortConfigs `yaml:"listen_addresses"`
}

// type check
var _ validator = (*serverConfig)(nil)

// validate implements the [validator] interface for *serverConfig.
func (c *serverConfig) validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	err = c.ListenAddresses.validate()
	if err != nil {
		return fmt.Errorf("listen_addresses: %w", err)
	}

	return nil
}
