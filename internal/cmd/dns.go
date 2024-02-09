package cmd

import (
	"fmt"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdc"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/mapsutil"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/timeutil"
	"github.com/c2h5oh/datasize"
)

// dnsConfig is the configuration for handling DNS.
type dnsConfig struct {
	// Cache configures the DNS results cache.
	Cache *cacheConfig `yaml:"cache"`

	// Server configures handling of incoming DNS requests.
	Server *servingConfig `yaml:"server"`

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
	defer func() { err = errors.Annotate(err, "dns section: %w") }()

	if c == nil {
		return errNoValue
	}

	validators := []validator{
		c.Cache,
		c.Server,
		c.Bootstrap,
		c.Upstream,
		c.Fallback,
	}

	var errs []error
	for _, v := range validators {
		errs = append(errs, v.validate())
	}

	return errors.Join(errs...)
}

// cacheConfig is the configuration for the DNS results cache.
type cacheConfig struct {
	// Enabled specifies if the cache should be used.
	Enabled bool `yaml:"enabled"`

	// Size is the maximum size of the cache.
	Size datasize.ByteSize `yaml:"size"`

	// ClientSize is the maximum size of the cache per client.
	ClientSize datasize.ByteSize `yaml:"client_size"`
}

// type check
var _ validator = (*cacheConfig)(nil)

// validate implements the [validator] interface for *cacheConfig.
func (c *cacheConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "cache section: %w") }()

	if c == nil {
		return errNoValue
	}

	var errs []error

	if c.Size <= 0 {
		errs = append(errs, fmt.Errorf("got size %s: %w", c.Size, errMustBeNonNegative))
	}

	if c.ClientSize <= 0 {
		errs = append(
			errs,
			fmt.Errorf("got client_size %s: %w", c.ClientSize, errMustBeNonNegative),
		)
	}

	return errors.Join(errs...)
}

// servingConfig is the configuration for serving DNS requests.
type servingConfig struct {
	// ListenAddresses is the addresses server listens for requests.
	ListenAddresses []*serverConfig `yaml:"listen_addresses"`
}

// type check
var _ validator = (*servingConfig)(nil)

// validate implements the [validator] interface for *servingConfig.
func (c *servingConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "server section: %w") }()

	switch {
	case c == nil:
		return errNoValue
	case len(c.ListenAddresses) == 0:
		return fmt.Errorf("listen_addresses: %w", errNoValue)
	}

	var errs []error

	for i, addr := range c.ListenAddresses {
		errs = append(errs, errors.Annotate(addr.validate(), "listen_addresses at index %d: %w", i))
	}

	return errors.Join(errs...)
}

// bootstrapConfig is the configuration for resolving upstream's hostnames.
type bootstrapConfig struct {
	// Servers is the list of DNS servers to use for resolving upstream's
	// hostnames.
	Servers []*serverConfig `yaml:"servers"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// type check
var _ validator = (*bootstrapConfig)(nil)

// validate implements the [validator] interface for *bootstrapConfig.
func (c *bootstrapConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "bootstrap section: %w") }()

	if c == nil {
		return errNoValue
	}

	var errs []error

	if c.Timeout.Duration <= 0 {
		errs = append(errs, fmt.Errorf("got timeout %s: %w", c.Timeout, errMustBePositive))
	}

	for i, s := range c.Servers {
		errs = append(errs, errors.Annotate(s.validate(), "servers at index %d: %w", i))
	}

	return errors.Join(errs...)
}

// fallbackConfig is the configuration for the fallback DNS upstream servers.
type fallbackConfig struct {
	// Servers is the list of DNS servers to use for fallback.
	Servers []*serverConfig `yaml:"servers"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// type check
var _ validator = (*fallbackConfig)(nil)

// validate implements the [validator] interface for *fallbackConfig.
func (c *fallbackConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "fallback section: %w") }()

	if c == nil {
		return errNoValue
	}

	var errs []error

	if c.Timeout.Duration <= 0 {
		errs = append(errs, fmt.Errorf("got timeout %s: %w", c.Timeout, errMustBePositive))
	}

	for i, s := range c.Servers {
		errs = append(errs, errors.Annotate(s.validate(), "servers at index %d: %w", i))
	}

	return errors.Join(errs...)
}

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

	if _, ok := c[agdc.UpstreamGroupNameDefault]; !ok {
		errs = append(errs, fmt.Errorf("group %q must be present", agdc.UpstreamGroupNameDefault))
	}

	mapsutil.OrderedRange(c, func(name agdc.UpstreamGroupName, g *upstreamGroupConfig) (cont bool) {
		err = g.validate()
		if err != nil {
			errs = append(errs, fmt.Errorf("group %q: %w", name, err))
		} else if name == agdc.UpstreamGroupNameDefault && g.Match != nil {
			errs = append(errs, fmt.Errorf("group %q must not have any match criteria", name))
		}

		return true
	})

	return errors.Join(errs...)
}

// serverConfig is the object for configuring an entity having an address.
//
// TODO(e.burkov):  Think more about naming, since it collides with the actual
// server section and doesn't really reflect the purpose of the object.
type serverConfig struct {
	// Address is the address of the server.
	//
	// TODO(e.burkov):  Perhaps, this should be more strictly typed.
	Address string `yaml:"address"`
}

// type check
var _ validator = (*serverConfig)(nil)

// validate implements the [validator] interface for *serverConfig.
//
// TODO(e.burkov):  Consider validating the address according to the particular
// configuration object's needs.
func (c *serverConfig) validate() (err error) {
	switch {
	case c == nil:
		return errNoValue
	case c.Address == "":
		return errEmptyValue
	default:
		return nil
	}
}

// upstreamGroupConfig is the configuration for a group of DNS upstream servers.
type upstreamGroupConfig struct {
	serverConfig `yaml:",inline"`

	// Match is the set of criteria for choosing this group.
	Match []*upstreamMatchConfig `yaml:"match"`
}

// type check
var _ validator = (*upstreamGroupConfig)(nil)

// validate implements the [validator] interface for *upstreamGroupConfig.
func (c *upstreamGroupConfig) validate() (err error) {
	if c == nil {
		return errNoValue
	}

	errs := []error{
		errors.Annotate(c.serverConfig.validate(), "server: %w"),
	}

	for i, m := range c.Match {
		errs = append(errs, errors.Annotate(m.validate(), "match at index %d: %w", i))
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
		return errors.Annotate(netutil.ValidateDomainName(c.QuestionDomain), "question_domain: %w")
	}

	return nil
}
