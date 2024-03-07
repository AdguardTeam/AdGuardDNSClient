package cmd

import (
	"fmt"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/timeutil"
)

// bootstrapConfig is the configuration for resolving upstream's hostnames.
type bootstrapConfig struct {
	// Servers is the list of DNS servers to use for resolving upstream's
	// hostnames.
	Servers ipPortAddressConfigs `yaml:"servers"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// type check
var _ validator = (*bootstrapConfig)(nil)

// validate implements the [validator] interface for *bootstrapConfig.
func (c *bootstrapConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "bootstrap section: %w") }()

	if c == nil {
		return errNoValue
	}

	var errs []error

	if c.Timeout.Duration <= 0 {
		err = fmt.Errorf("got timeout %s: %w", c.Timeout, errMustBePositive)

		errs = append(errs, err)
	}

	err = c.Servers.validate()
	if err != nil {
		err = fmt.Errorf("servers: %w", err)

		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
