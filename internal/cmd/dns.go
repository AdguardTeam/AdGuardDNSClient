package cmd

import (
	"fmt"
	"net/netip"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/netutil"
)

// dnsConfig is the configuration for handling DNS.
type dnsConfig struct {
	// Cache configures the DNS results cache.
	Cache *cacheConfig `yaml:"cache"`

	// Server configures handling of incoming DNS requests.
	Server *serverConfig `yaml:"server"`

	// Bootstrap configures the resolving of upstream's hostnames.
	Bootstrap *bootstrapConfig `yaml:"bootstrap"`

	// Upstream configures the DNS upstream servers.
	Upstream *upstreamConfig `yaml:"upstream"`

	// Fallback configures the fallback DNS upstream servers.
	Fallback *fallbackConfig `yaml:"fallback"`
}

// type check
var _ validator = (*dnsConfig)(nil)

// validate implements the [validator] interface for *dnsConfig.
func (c *dnsConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "dns: %w") }()

	if c == nil {
		return errNoValue
	}

	validators := []validator{
		c.Cache,
		c.Server,
		c.Bootstrap,
		c.Upstream,
		c.Fallback,
	}

	var errs []error
	for _, v := range validators {
		errs = append(errs, v.validate())
	}

	return errors.Join(errs...)
}

// toInternal converts the DNS configuration to the internal representation.  c
// must be valid.
func (c *dnsConfig) toInternal() (conf *dnssvc.Config) {
	return &dnssvc.Config{
		// TODO(e.burkov):  Consider making configurable.
		PrivateSubnets: netutil.SubnetSetFunc(netutil.IsLocallyServed),
		Cache:          c.Cache.toInternal(),
		Bootstrap:      c.Bootstrap.toInternal(),
		Upstreams:      c.Upstream.toInternal(),
		Fallbacks:      c.Fallback.toInternal(),
		ClientGetter:   dnssvc.DefaultClientGetter{},
		ListenAddrs:    c.Server.ListenAddresses.toInternal(),
	}
}

// ipPortConfig is the object for configuring an entity having an IP address
// with a port.
type ipPortConfig struct {
	// Address is the address of the server.
	Address netip.AddrPort `yaml:"address"`
}

// validate returns an error if c is not valid.  It doesn't include its own name
// into an error to be used in different configuration sections, and therefore
// violates the [validator.validate] contract.
func (c *ipPortConfig) validate() (err error) {
	switch {
	case c == nil:
		return errNoValue
	case c.Address == netip.AddrPort{}:
		return fmt.Errorf("address: %w", errEmptyValue)
	default:
		return nil
	}
}

// ipPortConfigs is a slice of *ipPortConfig for validation and conversion
// convenience.
type ipPortConfigs []*ipPortConfig

// toInternal converts the addresses to the internal representation.  c must be
// valid.
func (c ipPortConfigs) toInternal() (addrs []netip.AddrPort) {
	addrs = make([]netip.AddrPort, 0, len(c))
	for _, addr := range c {
		addrs = append(addrs, addr.Address)
	}

	return addrs
}

// validate returns an error if c is not valid.  It doesn't include its own name
// into an error to be used in different configuration sections, and therefore
// violates the [validator.validate] contract.
func (c ipPortConfigs) validate() (res error) {
	if len(c) == 0 {
		return errNoValue
	}

	var errs []error
	for i, addr := range c {
		err := addr.validate()
		if err != nil {
			err = fmt.Errorf("at index %d: %w", i, err)
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
