package cmd

import (
	"log/slog"
	"net/netip"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/container"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/validate"
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
var _ validate.Interface = (*dnsConfig)(nil)

// Validate implements the [validate.Interface] interface for *dnsConfig.
func (c *dnsConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	validators := container.KeyValues[string, validate.Interface]{{
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
		errs = validate.Append(errs, v.Key, v.Value)
	}

	return errors.Join(errs...)
}

// toInternal converts the DNS configuration to the internal representation.  c
// must be valid.
func (c *dnsConfig) toInternal(logger *slog.Logger) (conf *dnssvc.Config) {
	listenAddrs := make([]netip.AddrPort, 0, len(c.Server.ListenAddresses))
	for _, s := range c.Server.ListenAddresses {
		listenAddrs = append(listenAddrs, s.Address)
	}

	return &dnssvc.Config{
		BaseLogger: logger,
		Logger:     logger.With(slogutil.KeyPrefix, "dnssvc"),
		// TODO(e.burkov):  Consider making configurable.
		PrivateSubnets:  netutil.SubnetSetFunc(netutil.IsLocallyServed),
		Cache:           c.Cache.toInternal(),
		Bootstrap:       c.Bootstrap.toInternal(),
		Upstreams:       c.Upstream.toInternal(),
		Fallbacks:       c.Fallback.toInternal(),
		ClientGetter:    dnssvc.DefaultClientGetter{},
		ListenAddrs:     listenAddrs,
		BindRetry:       c.Server.BindRetry.toInternal(),
		PendingRequests: c.Server.PendingRequests.toInternal(),
	}
}

// ipPortConfig is the object for configuring an entity having an IP address
// with a port.
type ipPortConfig struct {
	// Address is the address of the server.
	Address netip.AddrPort `yaml:"address"`
}

// type check
var _ validate.Interface = (*ipPortConfig)(nil)

// Validate implements the [validate.Interface] interface for *ipPortConfig.
func (c *ipPortConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	return validate.NotEmpty("address", c.Address)
}
