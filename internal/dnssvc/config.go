package dnssvc

import (
	"fmt"
	"io"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdc"
	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/AdguardTeam/golibs/errors"
)

// Config is the configuration for [DNSService].
//
// TODO(e.burkov):  Add cache.
//
// TODO(e.burkov):  Add contracts.
type Config struct {
	// Bootstrap describes bootstrapping DNS servers.
	Bootstrap *BootstrapConfig

	// Upstreams describes DNS upstream servers.
	Upstreams *UpstreamConfig

	// Fallbacks describes DNS fallback upstream servers.
	Fallbacks *FallbackConfig

	// ListenAddrs is the list of served addresses.
	ListenAddrs []netip.AddrPort
}

// BootstrapConfig is the configuration for DNS bootstrap servers.
type BootstrapConfig struct {
	// Addresses is the list of servers.
	Addresses []netip.AddrPort

	// Timeout is the timeout for DNS requests.
	Timeout time.Duration
}

// toProxyResolvers creates a new bootstrap resolver and a list of upstreams to
// close on shutdown.
func (conf *BootstrapConfig) toProxyResolvers() (
	boot upstream.Resolver,
	closers []io.Closer,
	err error,
) {
	opts := &upstream.Options{
		Timeout: conf.Timeout,
	}

	resolvers := make(upstream.ConsequentResolver, 0, len(conf.Addresses))
	closers = make([]io.Closer, 0, len(conf.Addresses))
	var errs []error

	for i, addr := range conf.Addresses {
		var b *upstream.UpstreamResolver
		b, err = upstream.NewUpstreamResolver(addr.String(), opts)
		if err != nil {
			err = fmt.Errorf("creating bootstrap at index %d: %w", i, err)
			errs = append(errs, err)

			continue
		}

		resolvers = append(resolvers, upstream.NewCachingResolver(b))
		closers = append(closers, b.Upstream)
	}

	return slices.Clip(resolvers), slices.Clip(closers), errors.Join(errs...)
}

// UpstreamConfig is the configuration for DNS upstream servers.
type UpstreamConfig struct {
	// Groups is the list of groups.
	Groups []*UpstreamGroupConfig

	// Timeout is the timeout for DNS requests.
	Timeout time.Duration
}

// UpstreamGroupConfig is the configuration for a DNS upstream group.
type UpstreamGroupConfig struct {
	// Name is the name of the group.
	Name agdc.UpstreamGroupName

	// Address is the address of the server.
	Address string

	// Match is the list of match criteria.
	Match []MatchCriteria
}

// MatchCriteria is the criteria for matching the upstream group to handle DNS
// requests.
type MatchCriteria struct {
	// Client is the prefix to match the client address.
	Client netip.Prefix

	// QuestionDomain is the suffix to match the question domain.
	QuestionDomain string
}

// toProxyUpstreamConfig converts uc to a [proxy.UpstreamConfig], building it
// from the groups.
//
// TODO(e.burkov):  Use a more structured approach.
func (uc *UpstreamConfig) toProxyUpstreamConfig(
	opts *upstream.Options,
) (conf *proxy.UpstreamConfig, err error) {
	var addrs []string

	for _, group := range uc.Groups {
		line := group.toUpstreamConfigline()

		addrs = append(addrs, line)
	}

	return proxy.ParseUpstreamsConfig(addrs, opts)
}

// toUpstreamConfigline converts the group to a line for the
// [proxy.UpstreamConfig].
func (g *UpstreamGroupConfig) toUpstreamConfigline() (line string) {
	var domains []string

	switch g.Name {
	case agdc.UpstreamGroupNameDefault:
		return g.Address
	case agdc.UpstreamGroupNamePrivate:
		domains = privateDomains
	default:
		if len(g.Match) == 0 {
			return line
		}

		for _, match := range g.Match {
			if match.QuestionDomain != "" {
				domains = append(domains, match.QuestionDomain)
			}
		}
	}

	return fmt.Sprintf("[/%s/]%s", strings.Join(domains, "/"), g.Address)
}

// privateDomains returns the list of PTR domains considered private as per RFC
// 6303.
//
// TODO(e.burkov):  Get rid of it when proxy supports separate private upstream
// configuration.
var privateDomains = []string{
	"0.in-addr.arpa",
	"10.in-addr.arpa",
	"127.in-addr.arpa",
	"254.169.in-addr.arpa",
	"16.172.in-addr.arpa",
	"17.172.in-addr.arpa",
	"18.172.in-addr.arpa",
	"19.172.in-addr.arpa",
	"20.172.in-addr.arpa",
	"21.172.in-addr.arpa",
	"22.172.in-addr.arpa",
	"23.172.in-addr.arpa",
	"24.172.in-addr.arpa",
	"25.172.in-addr.arpa",
	"26.172.in-addr.arpa",
	"27.172.in-addr.arpa",
	"28.172.in-addr.arpa",
	"29.172.in-addr.arpa",
	"30.172.in-addr.arpa",
	"31.172.in-addr.arpa",
	"2.0.192.in-addr.arpa",
	"168.192.in-addr.arpa",
	"100.51.198.in-addr.arpa",
	"113.0.203.in-addr.arpa",
	"255.255.255.255.in-addr.arpa",
	"0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.ip6.arpa",
	"1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.ip6.arpa",
	"8.b.d.0.1.0.0.2.ip6.arpa",
	"d.f.ip6.arpa",
	"8.e.f.ip6.arpa",
	"9.e.f.ip6.arpa",
	"a.e.f.ip6.arpa",
	"b.e.f.ip6.arpa",
}

// FallbackConfig is the configuration for DNS fallback upstream servers.
type FallbackConfig struct {
	// Addresses is the list of servers.
	Addresses []string

	// Timeout is the timeout for DNS requests.
	Timeout time.Duration
}
