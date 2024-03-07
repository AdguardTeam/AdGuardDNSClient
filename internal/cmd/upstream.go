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

// type check
var _ validator = (*upstreamConfig)(nil)

// validate implements the [validator] interface for *upstreamConfig.
func (c *upstreamConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "upstream section: %w") }()

	if c == nil {
		return errNoValue
	}

	if c.Timeout.Duration <= 0 {
		err = fmt.Errorf("got timeout %s: %w", c.Timeout, errMustBePositive)
	}

	return errors.Join(err, c.Groups.validate())
}

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

	var errs []error
	for _, required := range []agdc.UpstreamGroupName{
		agdc.UpstreamGroupNameDefault,
		agdc.UpstreamGroupNamePrivate,
	} {
		if _, ok := c[required]; !ok {
			errs = append(errs, fmt.Errorf("group %q must be specified", required))
		}
	}

	mapsutil.OrderedRange(c, func(name agdc.UpstreamGroupName, g *upstreamGroupConfig) (cont bool) {
		err = g.validate(name)
		if err != nil {
			errs = append(errs, fmt.Errorf("group %q: %w", name, err))
		}

		return true
	})

	return errors.Join(errs...)
}

// upstreamGroupConfig is the configuration for a group of DNS upstream servers.
type upstreamGroupConfig struct {
	addressConfig `yaml:",inline"`

	// Match is the set of criteria for choosing this group.
	Match []*upstreamMatchConfig `yaml:"match"`
}

// validate implements the [validator] interface for *upstreamGroupConfig.
func (c *upstreamGroupConfig) validate(name agdc.UpstreamGroupName) (err error) {
	if c == nil {
		return errNoValue
	}

	var errs []error

	err = c.addressConfig.validate()
	if err != nil {
		errs = append(errs, fmt.Errorf("server: %w", err))
	}

	switch name {
	case
		agdc.UpstreamGroupNameDefault,
		agdc.UpstreamGroupNamePrivate:
		if len(c.Match) > 0 {
			errs = append(errs, errMustHaveNoMatch)
		}
	default:
		for i, m := range c.Match {
			err = m.validate()
			if err != nil {
				errs = append(errs, fmt.Errorf("match at index %d: %w", i, err))
			}
		}
	}

	return errors.Join(errs...)
}

// upstreamMatchConfig is the configuration for a criteria for choosing an
// upstream group.
type upstreamMatchConfig struct {
	// Client is the client's subnet to match.
	//
	// TODO(e.burkov):  Use.
	Client netutil.Prefix `yaml:"client"`

	// QuestionDomain is the domain name from request's question to match.
	QuestionDomain string `yaml:"question_domain"`
}

// type check
var _ validator = (*upstreamMatchConfig)(nil)

// validate implements the [validator] interface for *upstreamMatchConfig.
func (c *upstreamMatchConfig) validate() (err error) {
	if c == nil {
		return errNoValue
	} else if *c == (upstreamMatchConfig{}) {
		return errEmptyValue
	}

	if c.QuestionDomain != "" {
		err = netutil.ValidateDomainName(c.QuestionDomain)
		if err != nil {
			return fmt.Errorf("question_domain: %w", err)
		}
	}

	return nil
}
