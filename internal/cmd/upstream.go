package cmd

import (
	"fmt"
	"net/netip"
	"slices"
	"strings"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdc"
	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/mapsutil"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/timeutil"
)

// upstreamConfig is the configuration for the DNS upstream servers.
type upstreamConfig struct {
	// Groups contains all the groups of servers.
	Groups upstreamGroupsConfig `yaml:"groups"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// toInternal converts the configuration to a *dnssvc.UpstreamConfig.  c must be
// valid.
func (c *upstreamConfig) toInternal() (conf *dnssvc.UpstreamConfig) {
	conf = &dnssvc.UpstreamConfig{
		Timeout: c.Timeout.Duration,
	}
	rangeFunc := func(name agdc.UpstreamGroupName, g *upstreamGroupConfig) (cont bool) {
		grpConf := &dnssvc.UpstreamGroupConfig{
			Name:    name,
			Address: g.Address,
		}
		for _, m := range g.Match {
			grpConf.Match = append(grpConf.Match, dnssvc.MatchCriteria{
				Client:         m.Client.Prefix,
				QuestionDomain: m.QuestionDomain,
			})
		}

		conf.Groups = append(conf.Groups, grpConf)

		return true
	}

	mapsutil.OrderedRange(c.Groups, rangeFunc)

	return conf
}

// type check
var _ validator = (*upstreamConfig)(nil)

// validate implements the [validator] interface for *upstreamConfig.
func (c *upstreamConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "upstream: %w") }()

	if c == nil {
		return errNoValue
	}

	var errs []error

	if c.Timeout.Duration <= 0 {
		err = fmt.Errorf("got timeout %s: %w", c.Timeout, errMustBePositive)
		errs = append(errs, err)
	}

	if err = c.Groups.validate(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// indexedMatch is a key for matchSet.  It's essentially an
// [upstreamMatchConfig] with a lowercased question domain.
type indexedMatch struct {
	domain string
	client netip.Prefix
}

// matchSet validates that no two matches have the same domain and client in
// different upstream groups.
type matchSet map[indexedMatch]agdc.UpstreamGroupName

// validateMatches returns an error if the matches in g conflict with the ones
// in s.  name is the name of the group g.
func (s matchSet) validateMatches(g *upstreamGroupConfig, name agdc.UpstreamGroupName) (err error) {
	var errs []error
	for i, m := range g.Match {
		key := m.toIndexedMatch()

		another, ok := s[key]
		if !ok {
			s[key] = name

			continue
		}

		if another == name {
			err = errMustBeUnique
		} else {
			err = fmt.Errorf("conflicts with group %q", another)
		}

		err = fmt.Errorf("match: at index %d: %w", i, err)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// upstreamGroupsConfig is the configuration for a set of groups of DNS upstream
// servers.
type upstreamGroupsConfig map[agdc.UpstreamGroupName]*upstreamGroupConfig

// requiredGroups is the list of groups that must be present in a valid
// [upstreamGroupsConfig].  Those should also have no match criteria.
//
// TODO(e.burkov):  Add IsRequired method to UpstreamGroupName?
var requiredGroups = []agdc.UpstreamGroupName{
	agdc.UpstreamGroupNameDefault,
	// TODO(e.burkov):  Add UpstreamGroupNamePrivate.
}

// type check
var _ validator = (upstreamGroupsConfig)(nil)

// validate implements the [validator] interface for upstreamGroupsConfig.
func (c upstreamGroupsConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "groups: %w") }()

	if c == nil {
		return errNoValue
	}

	errs := c.validateRequiredGroups()
	errs = append(errs, c.validateGroups()...)

	return errors.Join(errs...)
}

// validateRequiredGroups returns errors of validating the required groups
// within c.
func (c upstreamGroupsConfig) validateRequiredGroups() (errs []error) {
	for _, required := range requiredGroups {
		if g, ok := c[required]; !ok {
			errs = append(errs, fmt.Errorf("group %q must be specified", required))
		} else if len(g.Match) > 0 {
			errs = append(errs, fmt.Errorf("group %q: %w", required, errMustHaveNoMatch))
		}
	}

	return errs
}

// validateGroups returns errors of validating the groups within c.
func (c upstreamGroupsConfig) validateGroups() (errs []error) {
	matches := matchSet{}
	mapsutil.OrderedRange(c, func(name agdc.UpstreamGroupName, g *upstreamGroupConfig) (cont bool) {
		err := g.validate()
		if err == nil && !slices.Contains(requiredGroups, name) {
			// Only validate matches if the group is valid and is expected to
			// have them.
			err = matches.validateMatches(g, name)
		}

		if err != nil {
			err = fmt.Errorf("group %q: %w", name, err)
			errs = append(errs, err)
		}

		return true
	})

	return errs
}

// upstreamGroupConfig is the configuration for a group of DNS upstream servers.
type upstreamGroupConfig struct {
	// Address is the URL of the upstream server for this group.
	Address string `yaml:"address"`

	// Match is the set of criteria for choosing this group.
	Match []*upstreamMatchConfig `yaml:"match"`
}

// validate returns an error if c is not valid.  It doesn't include its own name
// into an error to be wrapped with different group names, and therefore
// violates the [validator.validate] contract.
func (c *upstreamGroupConfig) validate() (err error) {
	if c == nil {
		return errNoValue
	}

	var errs []error

	if c.Address == "" {
		errs = append(errs, fmt.Errorf("address: %w", errNoValue))
	}

	for i, m := range c.Match {
		err = m.validateValues()
		if err != nil {
			err = fmt.Errorf("match: at index %d: %w", i, err)
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// upstreamMatchConfig is the configuration for criteria for choosing an
// upstream group.
type upstreamMatchConfig struct {
	// Client is the client's subnet to match.  Prefix itself should be masked.
	Client netutil.Prefix `yaml:"client"`

	// QuestionDomain is the domain name from request's question to match.
	QuestionDomain string `yaml:"question_domain"`
}

// validateValues returns an error if c contains invalid question domain or
// client's prefix.
func (c *upstreamMatchConfig) validateValues() (err error) {
	switch {
	case c == nil:
		return errNoValue
	case *c == (upstreamMatchConfig{}):
		return errEmptyValue
	default:
		// Go on.
	}

	var errs []error

	if c.QuestionDomain != "" {
		err = netutil.ValidateDomainName(c.QuestionDomain)
		if err != nil {
			err = fmt.Errorf("question_domain: %w", err)
			errs = append(errs, err)
		}
	}

	if c.Client.Prefix != c.Client.Masked() {
		err = fmt.Errorf("client: %s must has %d significant bits", c.Client, c.Client.Bits())
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// toIndexedMatch converts the upstream match configuration to a key for
// [matchSet].
func (c *upstreamMatchConfig) toIndexedMatch() (im indexedMatch) {
	return indexedMatch{
		domain: strings.ToLower(c.QuestionDomain),
		client: c.Client.Prefix,
	}
}
