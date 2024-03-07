package cmd

import (
	"fmt"
	"net/netip"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/errors"
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
	defer func() { err = errors.Annotate(err, "dns section: %w") }()

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

// serverConfig is the configuration for serving DNS requests.
type serverConfig struct {
	// ListenAddresses is the addresses server listens for requests.
	ListenAddresses ipPortAddressConfigs `yaml:"listen_addresses"`
}

// type check
var _ validator = (*serverConfig)(nil)

// validate implements the [validator] interface for *serverConfig.
func (c *serverConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "server section: %w") }()

	switch {
	case c == nil:
		return errNoValue
	}

	err = c.ListenAddresses.validate()
	if err != nil {
		return fmt.Errorf("listen_addresses: %w", err)
	}

	return nil
}

// addressConfig is the object for configuring an entity having an address.
//
// TODO(e.burkov):  Think more about naming, since it collides with the actual
// server section and doesn't really reflect the purpose of the object.
type addressConfig struct {
	// Address is the address of the server.
	//
	// TODO(e.burkov):  Perhaps, this should be more strictly typed.
	Address string `yaml:"address"`
}

// type check
var _ validator = (*addressConfig)(nil)

// validate implements the [validator] interface for *addressConfig.
func (c *addressConfig) validate() (err error) {
	switch {
	case c == nil:
		return errNoValue
	case c.Address == "":
		return errEmptyValue
	default:
		return nil
	}
}

// ipAddressConfigs is a slice of *addressConfig that should be IP addresses
// with ports.
type ipPortAddressConfigs []*addressConfig

// type check
var _ validator = (ipPortAddressConfigs)(nil)

// validate implements the [validator] interface for ipAddressConfigs.
func (c ipPortAddressConfigs) validate() (err error) {
	if len(c) == 0 {
		return errNoValue
	}

	var errs []error
	for i, addr := range c {
		err = addr.validate()
		if err == nil {
			_, err = netip.ParseAddrPort(addr.Address)
		}

		if err != nil {
			errs = append(errs, fmt.Errorf("address at index %d: %w", i, err))
		}
	}

	return errors.Join(errs...)
}

// urlAddressCOnfigs is a slice of *addressConfig that should be URLs.
type urlAddressConfigs []*addressConfig

// type check
var _ validator = (urlAddressConfigs)(nil)

// validate implements the [validator] interface for urlAddressConfigs.
func (c urlAddressConfigs) validate() (err error) {
	if len(c) == 0 {
		return errNoValue
	}

	var errs []error
	for i, addr := range c {
		err = addr.validate()
		if err != nil {
			errs = append(errs, fmt.Errorf("address at index %d: %w", i, err))
		}
	}

	return errors.Join(errs...)
}

// toInternal converts the DNS configuration to the internal representation.
func (c *dnsConfig) toInternal() (conf *dnssvc.Config, err error) {
	listenAddrs := make([]netip.AddrPort, 0, len(c.Server.ListenAddresses))
	for i, addr := range c.Server.ListenAddresses {
		var addrPort netip.AddrPort
		addrPort, err = netip.ParseAddrPort(addr.Address)
		if err != nil {
			return nil, fmt.Errorf("listen address at index %d: %w", i, err)
		}

		listenAddrs = append(listenAddrs, addrPort)
	}

	bootstraps := make([]netip.AddrPort, 0, len(c.Bootstrap.Servers))
	for i, s := range c.Bootstrap.Servers {
		var addrPort netip.AddrPort
		addrPort, err = netip.ParseAddrPort(s.Address)
		if err != nil {
			return nil, fmt.Errorf("bootstrap server at index %d: %w", i, err)
		}

		bootstraps = append(bootstraps, addrPort)
	}

	falls := make([]string, 0, len(c.Fallback.Servers))
	for _, s := range c.Fallback.Servers {
		falls = append(falls, s.Address)
	}

	return &dnssvc.Config{
		ListenAddrs: listenAddrs,
		Bootstrap: &dnssvc.BootstrapConfig{
			Addresses: bootstraps,
			Timeout:   c.Bootstrap.Timeout.Duration,
		},
		Upstreams: c.Upstream.toInternal(),
		Fallbacks: &dnssvc.FallbackConfig{
			Addresses: falls,
			Timeout:   c.Fallback.Timeout.Duration,
		},
	}, nil
}
