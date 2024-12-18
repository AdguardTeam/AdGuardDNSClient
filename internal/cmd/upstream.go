package cmd

import (
	"fmt"
	"maps"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdc"
	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/timeutil"
	"github.com/AdguardTeam/golibs/validate"
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
		Timeout: time.Duration(c.Timeout),
	}

	for name, g := range c.Groups {
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
	}

	return conf
}

// type check
var _ validate.Interface = (*upstreamConfig)(nil)

// Validate implements the [validate.Interface] interface for *upstreamConfig.
func (c *upstreamConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	errs := []error{
		validate.Positive("timeout", c.Timeout),
	}
	errs = validate.Append(errs, "groups", c.Groups)

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

// addMatch returns an error if m conflicts with the ones in s.  name is the
// name of the group containing m.
func (s matchSet) addMatch(name agdc.UpstreamGroupName, m *upstreamMatchConfig) (err error) {
	key := m.toIndexedMatch()
	another, ok := s[key]
	if !ok {
		s[key] = name

		return nil
	}

	if another == name {
		return errors.ErrDuplicated
	}

	return fmt.Errorf("conflicts with group %q", another)
}

// upstreamGroupsConfig is the configuration for a set of groups of DNS upstream
// servers.
type upstreamGroupsConfig map[agdc.UpstreamGroupName]*upstreamGroupConfig

// requiredGroups is the list of groups that must be present in a valid
// [upstreamGroupsConfig].
var requiredGroups = []agdc.UpstreamGroupName{
	agdc.UpstreamGroupNameDefault,
}

// predefinedGroups is the list of groups that must have no match criteria in a
// valid [upstreamGroupsConfig].
var predefinedGroups = []agdc.UpstreamGroupName{
	agdc.UpstreamGroupNameDefault,
	agdc.UpstreamGroupNamePrivate,
}

// type check
var _ validate.Interface = (upstreamGroupsConfig)(nil)

// Validate implements the [validate.Interface] interface for
// upstreamGroupsConfig.
func (c upstreamGroupsConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	var errs []error
	for _, name := range requiredGroups {
		if _, ok := c[name]; !ok {
			err = fmt.Errorf("group %q: must be present", name)
			errs = append(errs, err)
		}
	}

	errs = c.validateGroups(errs)

	return errors.Join(errs...)
}

// validateGroups appends the errors of validating groups within c to errs and
// returns the result.
func (c upstreamGroupsConfig) validateGroups(errs []error) (res []error) {
	ms := matchSet{}
	for _, name := range slices.Sorted(maps.Keys(c)) {
		g := c[name]

		var err error
		if slices.Contains(predefinedGroups, name) {
			err = g.validateAsPredefined()
		} else {
			err = g.validateAsCustom(ms, name)
		}
		if err != nil {
			err = fmt.Errorf("group %q: %w", name, err)
			errs = append(errs, err)
		}
	}

	return errs
}

// upstreamGroupConfig is the configuration for a group of DNS upstream servers.
type upstreamGroupConfig struct {
	// Address is the URL of the upstream server for this group.
	Address string `yaml:"address"`

	// Match is the set of criteria for choosing this group.
	Match []*upstreamMatchConfig `yaml:"match"`
}

// validateAsPredefined returns an error if c is not a valid predefined group
// configuration that should have no match criteria.
func (c *upstreamGroupConfig) validateAsPredefined() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	errs := []error{
		validate.NotEmpty("address", c.Address),
	}

	// TODO(e.burkov):  Add validate.Empty for entities that must be empty.
	if len(c.Match) > 0 {
		err = fmt.Errorf("match: %w", errors.ErrNotEmpty)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// validateAsCustom returns an error if c is not a valid custom group
// configuration for group named n within the set s.
func (c *upstreamGroupConfig) validateAsCustom(s matchSet, n agdc.UpstreamGroupName) (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	errs := []error{
		validate.NotEmpty("address", c.Address),
	}

	for i, m := range c.Match {
		err = m.validate(s, n)
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

// validate returns error if c is not valid.
func (c *upstreamMatchConfig) validate(s matchSet, name agdc.UpstreamGroupName) (err error) {
	switch {
	case c == nil:
		return errors.ErrNoValue
	case *c == (upstreamMatchConfig{}):
		return errors.ErrEmptyValue
	default:
		return c.validateValues(s, name)
	}
}

// validateValues returns error if c contains invalid values.  c must not be
// nil.
func (c *upstreamMatchConfig) validateValues(s matchSet, name agdc.UpstreamGroupName) (err error) {
	var errs []error

	if c.QuestionDomain != "" {
		err = netutil.ValidateDomainName(c.QuestionDomain)
		if err != nil {
			err = fmt.Errorf("question_domain: %w", err)
			errs = append(errs, err)
		}
	}

	// TODO(e.burkov):  It may be useful to be able to specify the whole address
	// and only change the mask.
	if c.Client.Prefix != c.Client.Masked() {
		bitNum := c.Client.Bits()
		err = fmt.Errorf("client: %s must has at most %d significant bits", c.Client, bitNum)
		errs = append(errs, err)
	}

	errs = append(errs, s.addMatch(name, c))

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
