package cmd

import (
	"fmt"
	"log/slog"
	"net/netip"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/container"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
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
	if c == nil {
		return errors.ErrNoValue
	}

	validators := container.KeyValues[string, validator]{{
		Key:   "cache",
		Value: c.Cache,
	}, {
		Key:   "server",
		Value: c.Server,
	}, {
		Key:   "bootstrap",
		Value: c.Bootstrap,
	}, {
		Key:   "upstream",
		Value: c.Upstream,
	}, {
		Key:   "fallback",
		Value: c.Fallback,
	}}

	var errs []error
	for _, v := range validators {
		err = v.Value.validate()
		if err != nil {
			err = fmt.Errorf("%s: %w", v.Key, err)
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// toInternal converts the DNS configuration to the internal representation.  c
// must be valid.
func (c *dnsConfig) toInternal(logger *slog.Logger) (conf *dnssvc.Config) {
	return &dnssvc.Config{
		BaseLogger: logger,
		Logger:     logger.With(slogutil.KeyPrefix, "dnssvc"),
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

// type check
var _ validator = (*ipPortConfig)(nil)

// validate implements the [validator] interface for *ipPortConfig.
func (c *ipPortConfig) validate() (err error) {
	switch {
	case c == nil:
		return errors.ErrNoValue
	case c.Address == netip.AddrPort{}:
		return fmt.Errorf("address: %w", errors.ErrEmptyValue)
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

// type check
var _ validator = (ipPortConfigs)(nil)

// validate implements the [validator] interface for ipPortConfigs.
func (c ipPortConfigs) validate() (res error) {
	if len(c) == 0 {
		return errors.ErrNoValue
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
