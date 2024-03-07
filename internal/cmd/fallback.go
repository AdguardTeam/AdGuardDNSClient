package cmd

import (
	"fmt"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/timeutil"
)

// fallbackConfig is the configuration for the fallback DNS upstream servers.
type fallbackConfig struct {
	// Servers is the list of DNS servers to use for fallback.
	Servers urlAddressConfigs `yaml:"servers"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// type check
var _ validator = (*fallbackConfig)(nil)

// validate implements the [validator] interface for *fallbackConfig.
func (c *fallbackConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "fallback section: %w") }()

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
