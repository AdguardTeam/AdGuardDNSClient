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
) (general *proxy.UpstreamConfig, err error) {
	opts := &upstream.Options{
		Timeout:   conf.Timeout,
		Bootstrap: boot,
	}

	config := &proxy.UpstreamConfig{}
	upstreams := map[string]upstream.Upstream{}

	var errs []error
	for _, g := range conf.Groups {
		err = g.addGroup(config, upstreams, opts)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("creating upstreams: %w", errors.Join(errs...))
	}

	return config, nil
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

// addGroup creates and puts the upstream from ugc into conf.  addrToUps is used
// to avoid creating upstreams from the identical addresses.  opts are used for
// creating upstreams.
func (ugc *UpstreamGroupConfig) addGroup(
	conf *proxy.UpstreamConfig,
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
		conf.Upstreams = append(conf.Upstreams, u)

		return nil
	}

	for _, m := range ugc.Match {
		addUpstream(conf, u, strings.ToLower(m.QuestionDomain))
	}

	return nil
}

// addUpstream adds u to the conf.  u considered a domain-specific upstream, if
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

	domain = dns.Fqdn(domain)
	conf.DomainReservedUpstreams[domain] = append(conf.DomainReservedUpstreams[domain], u)
	conf.SpecifiedDomainUpstreams[domain] = append(conf.SpecifiedDomainUpstreams[domain], u)
}

// MatchCriteria is the criteria for matching the upstream group to handle DNS
// requests.
type MatchCriteria struct {
	// Client is the prefix to match the client address.
	Client netip.Prefix

	// QuestionDomain is the suffix to match the question domain.
	QuestionDomain string
}
