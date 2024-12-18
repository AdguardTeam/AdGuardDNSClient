package cmd

import (
	"net/netip"
	"time"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/timeutil"
	"github.com/AdguardTeam/golibs/validate"
)

// bootstrapConfig is the configuration for resolving upstream's hostnames.
type bootstrapConfig struct {
	// Servers is the list of DNS servers to use for resolving upstream's
	// hostnames.
	Servers []*ipPortConfig `yaml:"servers"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// toInternal converts the bootstrap configuration to the internal
// representation.  c must be valid.
func (c *bootstrapConfig) toInternal() (conf *dnssvc.BootstrapConfig) {
	addrs := make([]netip.AddrPort, 0, len(c.Servers))
	for _, s := range c.Servers {
		addrs = append(addrs, s.Address)
	}

	return &dnssvc.BootstrapConfig{
		Timeout:   time.Duration(c.Timeout),
		Addresses: addrs,
	}
}

// type check
var _ validate.Interface = (*bootstrapConfig)(nil)

// Validate implements the [validate.Interface] interface for *bootstrapConfig.
func (c *bootstrapConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	errs := []error{
		validate.Positive("timeout", c.Timeout),
	}
	errs = validate.AppendSlice(errs, "servers", c.Servers)

	return errors.Join(errs...)
}
