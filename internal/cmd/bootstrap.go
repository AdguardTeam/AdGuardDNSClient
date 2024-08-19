package cmd

import (
	"fmt"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/timeutil"
)

// bootstrapConfig is the configuration for resolving upstream's hostnames.
type bootstrapConfig struct {
	// Servers is the list of DNS servers to use for resolving upstream's
	// hostnames.
	Servers ipPortConfigs `yaml:"servers"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// toInternal converts the bootstrap configuration to the internal
// representation.  c must be valid.
func (c *bootstrapConfig) toInternal() (conf *dnssvc.BootstrapConfig) {
	return &dnssvc.BootstrapConfig{
		Timeout:   c.Timeout.Duration,
		Addresses: c.Servers.toInternal(),
	}
}

// type check
var _ validator = (*bootstrapConfig)(nil)

// validate implements the [validator] interface for *bootstrapConfig.
func (c *bootstrapConfig) validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	var errs []error

	if c.Timeout.Duration <= 0 {
		err = fmt.Errorf("got timeout %s: %w", c.Timeout, errors.ErrNotPositive)
		errs = append(errs, err)
	}

	err = c.Servers.validate()
	if err != nil {
		err = fmt.Errorf("servers: %w", err)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
