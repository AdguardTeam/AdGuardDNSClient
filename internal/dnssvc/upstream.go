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
// TODO(e.burkov):  Put the default groups into separate fields.
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
) (configs map[netip.Prefix]*proxy.UpstreamConfig, err error) {
	defer func() { err = errors.Annotate(err, "creating upstreams: %w") }()

	opts := &upstream.Options{
		Timeout:   conf.Timeout,
		Bootstrap: boot,
	}

	configs = map[netip.Prefix]*proxy.UpstreamConfig{}
	upstreams := map[string]upstream.Upstream{}

	var errs []error
	for _, g := range conf.Groups {
		err = g.addGroup(configs, upstreams, opts)
		if err != nil {
			err = fmt.Errorf("adding group %q: %w", g.Name, err)
			errs = append(errs, err)
		}
	}

	return configs, errors.Join(errs...)
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

// addGroup creates and puts the upstream from ugc into configs.  addrToUps is
// used to avoid creating upstreams from the identical addresses.  opts are used
// for creating upstreams.
//
// TODO(e.burkov):  Lowercase addrs.
func (ugc *UpstreamGroupConfig) addGroup(
	configs map[netip.Prefix]*proxy.UpstreamConfig,
	addrToUps map[string]upstream.Upstream,
	opts *upstream.Options,
) (err error) {
	u, ok := addrToUps[ugc.Address]
	if !ok {
		u, err = upstream.AddressToUpstream(ugc.Address, opts)
		if err != nil {
			return err
		}

		addrToUps[ugc.Address] = u
	}

	if ugc.Name == agdc.UpstreamGroupNameDefault {
		configs[netip.Prefix{}] = &proxy.UpstreamConfig{
			Upstreams: []upstream.Upstream{u},
		}

		return nil
	}

	for _, m := range ugc.Match {
		conf := configs[m.Client]
		if conf == nil {
			conf = &proxy.UpstreamConfig{}
			configs[m.Client] = conf
		}

		addUpstream(conf, u, m.QuestionDomain)
	}

	return nil
}

// addUpstream adds u to conf.  u is considered a domain-specific upstream, if
// domain is not empty.
func addUpstream(conf *proxy.UpstreamConfig, u upstream.Upstream, domain string) {
	if domain == "" {
		conf.Upstreams = append(conf.Upstreams, u)

		return
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
