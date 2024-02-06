package cmd

import (
	"github.com/AdguardTeam/AdGuardDNSClient/internal/agc"
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

// cacheConfig is the configuration for the DNS results cache.
type cacheConfig struct {
	// Enabled specifies if the cache should be used.
	Enabled bool `yaml:"enabled"`

	// Size is the maximum size of the cache.
	Size datasize.ByteSize `yaml:"size"`

	// ClientSize is the maximum size of the cache per client.
	ClientSize datasize.ByteSize `yaml:"client_size"`
}

// servingConfig is the configuration for serving DNS requests.
type servingConfig struct {
	// ListenAddresses is the addresses server listens for requests.
	ListenAddresses []*serverConfig `yaml:"listen_addresses"`
}

// bootstrapConfig is the configuration for resolving upstream's hostnames.
type bootstrapConfig struct {
	// Servers is the list of DNS servers to use for resolving upstream's
	// hostnames.
	Servers []*serverConfig `yaml:"servers"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// upstreamConfig is the configuration for the DNS upstream servers.
type upstreamConfig struct {
	// Groups contains all the grous of servers.
	Groups map[agc.UpstreamGroupName]*upstreamGroupConfig `yaml:"groups"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// fallbackConfig is the configuration for the fallback DNS upstream servers.
type fallbackConfig struct {
	// Servers is the list of DNS servers to use for fallback.
	Servers []*serverConfig `yaml:"servers"`

	// Timeout constrains the time for sending requests and receiving responses.
	Timeout timeutil.Duration `yaml:"timeout"`
}

// serverConfig is the configuration for a DNS server.
type serverConfig struct {
	// Address is the address of the server.
	//
	// TODO(e.burkov):  Perhaps, this should be more strictly typed.
	Address string `yaml:"address"`
}

// upstreamGroupConfig is the configuration for a group of DNS upstream servers.
type upstreamGroupConfig struct {
	serverConfig `yaml:",inline"`

	// Match is the set of criteria for choosing this group.
	Match []*upstreamMatchConfig `yaml:"match"`
}

// upstreamMatchConfig is the configuration for a criteria for choosing an
// upstream group.
type upstreamMatchConfig struct {
	// Client is the client's subnet to match.
	Client netutil.Prefix `yaml:"client"`

	// QuestionDomain is the domain name from request's question to match.
	QuestionDomain string `yaml:"question_domain"`
}
