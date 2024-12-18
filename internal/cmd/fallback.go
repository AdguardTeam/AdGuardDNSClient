package cmd

import (
	"time"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/timeutil"
	"github.com/AdguardTeam/golibs/validate"
)

// fallbackConfig is the configuration for the fallback DNS upstream servers.
type fallbackConfig struct {
	// Servers is the list of DNS servers to use for fallback.
	Servers []*urlConfig `yaml:"servers"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// type check
var _ validate.Interface = (*fallbackConfig)(nil)

// Validate implements the [validate.Interface] interface for *fallbackConfig.
func (c *fallbackConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	errs := []error{
		validate.Positive("timeout", c.Timeout),
	}
	errs = validate.AppendSlice(errs, "servers", c.Servers)

	return errors.Join(errs...)
}

// toInternal converts the configuration to a *dnssvc.FallbackConfig.  c must be
// valid.
func (c *fallbackConfig) toInternal() (conf *dnssvc.FallbackConfig) {
	conf = &dnssvc.FallbackConfig{
		Timeout: time.Duration(c.Timeout),
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

// type check
var _ validate.Interface = (*urlConfig)(nil)

// Validate implements the [validate.Interface] interface for *urlConfig.
func (c *urlConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	return validate.NotEmpty("address", c.Address)
}
