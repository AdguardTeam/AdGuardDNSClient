package cmd

import (
	"fmt"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/timeutil"
)

// fallbackConfig is the configuration for the fallback DNS upstream servers.
type fallbackConfig struct {
	// Servers is the list of DNS servers to use for fallback.
	Servers urlConfigs `yaml:"servers"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// type check
var _ validator = (*fallbackConfig)(nil)

// validate implements the [validator] interface for *fallbackConfig.
func (c *fallbackConfig) validate() (err error) {
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

// toInternal converts the configuration to a *dnssvc.FallbackConfig.  c must be
// valid.
func (c *fallbackConfig) toInternal() (conf *dnssvc.FallbackConfig) {
	conf = &dnssvc.FallbackConfig{
		Timeout: c.Timeout.Duration,
	}

	for _, addrConf := range c.Servers {
		conf.Addresses = append(conf.Addresses, addrConf.Address)
	}

	return conf
}

// urlConfig is the object for configuring an entity having a URL address.
type urlConfig struct {
	// Address is the address of the server.
	Address string `yaml:"address"`
}

// urlConfigs is a slice of *urlConfig for validation convenience.
type urlConfigs []*urlConfig

// validate returns an error if c is not valid.  It doesn't include its own name
// into an error to be used in different configuration sections, and therefore
// violates the [validator.validate] contract.
func (c urlConfigs) validate() (err error) {
	if len(c) == 0 {
		return errors.ErrNoValue
	}

	var errs []error

	for i, addr := range c {
		switch {
		case addr == nil:
			err = errors.ErrNoValue
		case addr.Address == "":
			err = errors.ErrEmptyValue
		default:
			continue
		}

		err = fmt.Errorf("at index %d: address: %w", i, err)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
