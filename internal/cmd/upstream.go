package cmd

import (
	"fmt"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdc"
	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/mapsutil"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/timeutil"
)

// upstreamConfig is the configuration for the DNS upstream servers.
type upstreamConfig struct {
	// Groups contains all the grous of servers.
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

// upstreamGroupsConfig is the configuration for the set of groups of DNS
// upstream servers.
type upstreamGroupsConfig map[agdc.UpstreamGroupName]*upstreamGroupConfig

// type check
var _ validator = (upstreamGroupsConfig)(nil)

// validate implements the [validator] interface for upstreamGroupsConfig.
func (c upstreamGroupsConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "groups: %w") }()

	if c == nil {
		return errNoValue
	} else if len(c) == 0 {
		return errEmptyValue
	}

	errs := c.validateRequiredGroups()

	errs = append(errs, c.validateGroups()...)

	return errors.Join(errs...)
}

// validateRequiredGroups returns errors of validating the required groups
// within c.
func (c upstreamGroupsConfig) validateRequiredGroups() (errs []error) {
	for _, required := range []agdc.UpstreamGroupName{
		agdc.UpstreamGroupNameDefault,
		// TODO(e.burkov):  Add UpstreamGroupNamePrivate.
	} {
		if g, ok := c[required]; !ok {
			errs = append(errs, fmt.Errorf("group %q must be specified", required))
		} else if len(g.Match) > 0 {
			errs = append(errs, fmt.Errorf("group %q: %w", required, errMustHaveNoMatch))
		}
	}

	return errs
}

// validateGroups returns errors of validating the groups within c.
//
// TODO(e.burkov):  Skip required groups and require matches.
func (c upstreamGroupsConfig) validateGroups() (errs []error) {
	mapsutil.OrderedRange(c, func(name agdc.UpstreamGroupName, g *upstreamGroupConfig) (cont bool) {
		err := g.validate()
		if err != nil {
			errs = append(errs, fmt.Errorf("group %q: %w", name, err))
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
		err = m.validate()
		if err != nil {
			err = fmt.Errorf("match: at index %d: %w", i, err)

			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// upstreamMatchConfig is the configuration for a criteria for choosing an
// upstream group.
type upstreamMatchConfig struct {
	// Client is the client's subnet to match.
	Client netutil.Prefix `yaml:"client"`

	// QuestionDomain is the domain name from request's question to match.
	QuestionDomain string `yaml:"question_domain"`
}

// validate returns an error if c is not valid.  It doesn't include its own name
// into an error to be used in different configuration sections, and therefore
// violates the [validator.validate] contract.
func (c *upstreamMatchConfig) validate() (err error) {
	if c == nil {
		return errNoValue
	} else if *c == (upstreamMatchConfig{}) {
		return errEmptyValue
	}

	if c.QuestionDomain == "" {
		return nil
	}

	err = netutil.ValidateDomainName(c.QuestionDomain)
	if err != nil {
		return fmt.Errorf("question_domain: %w", err)
	}

	return nil
}
