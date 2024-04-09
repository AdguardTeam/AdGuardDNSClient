package dnssvc

import (
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdc"
	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/miekg/dns"
)

// UpstreamConfig is the configuration for DNS upstream servers.
//
// TODO(e.burkov):  Put the required groups into separate fields.
type UpstreamConfig struct {
	// Groups is the list of groups.
	Groups []*UpstreamGroupConfig

	// Timeout is the timeout for DNS requests.
	Timeout time.Duration
}

// newUpstreams converts conf to a [proxy.UpstreamConfig], building it from the
// groups.
func newUpstreams(
	conf *UpstreamConfig,
	boot upstream.Resolver,
) (ups upstreamConfigs, private *proxy.UpstreamConfig, err error) {
	defer func() { err = errors.Annotate(err, "creating upstreams: %w") }()

	opts := &upstream.Options{
		Timeout:   conf.Timeout,
		Bootstrap: boot,
	}

	private = &proxy.UpstreamConfig{}
	ups = upstreamConfigs{
		// Init default group.
		netip.Prefix{}: &proxy.UpstreamConfig{},
	}
	upstreams := map[string]upstream.Upstream{}

	var errs []error
	for _, g := range conf.Groups {
		var u upstream.Upstream
		u, err = newUpstreamOrCached(g.Address, upstreams, opts)
		if err != nil {
			errs = append(errs, fmt.Errorf("group %q: %w", g.Name, err))

			continue
		}

		switch g.Name {
		case agdc.UpstreamGroupNameDefault:
			ups[netip.Prefix{}].Upstreams = append(ups[netip.Prefix{}].Upstreams, u)
		case agdc.UpstreamGroupNamePrivate:
			private.Upstreams = append(private.Upstreams, u)
		default:
			g.addGroup(ups, u)
		}
	}

	return ups, private, errors.Join(errs...)
}

// newUpstreamOrCached creates a new upstream or returns the cached one from
// addrToUps.
func newUpstreamOrCached(
	addr string,
	addrToUps map[string]upstream.Upstream,
	opts *upstream.Options,
) (u upstream.Upstream, err error) {
	u, ok := addrToUps[addr]
	if !ok {
		u, err = upstream.AddressToUpstream(addr, opts)
		if err != nil {
			// Don't wrap the error, because it's informative enough as is.
			return nil, err
		}

		addrToUps[addr] = u
	}

	return u, nil
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

// addGroup adds u to the configuration of the corresponding client.
func (ugc *UpstreamGroupConfig) addGroup(configs upstreamConfigs, u upstream.Upstream) {
	for _, m := range ugc.Match {
		conf := configs[m.Client]
		if conf == nil {
			conf = &proxy.UpstreamConfig{}
			configs[m.Client] = conf
		}

		domain := m.QuestionDomain
		if domain == "" {
			conf.Upstreams = append(conf.Upstreams, u)

			continue
		}

		if conf.DomainReservedUpstreams == nil {
			conf.DomainReservedUpstreams = map[string][]upstream.Upstream{}
		}
		if conf.SpecifiedDomainUpstreams == nil {
			conf.SpecifiedDomainUpstreams = map[string][]upstream.Upstream{}
		}

		domain = dns.Fqdn(strings.ToLower(domain))
		conf.DomainReservedUpstreams[domain] = append(conf.DomainReservedUpstreams[domain], u)
		conf.SpecifiedDomainUpstreams[domain] = append(conf.SpecifiedDomainUpstreams[domain], u)
	}
}
