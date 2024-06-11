package dnssvc

import (
	"net/netip"

	"github.com/AdguardTeam/golibs/netutil"
)

// Config is the configuration for [DNSService].
//
// TODO(e.burkov):  Add cache.
type Config struct {
	// PrivateSubnets is the set of IP networks considered private.  The PTR
	// requests for ARPA domains considered private if the domain contains an IP
	// from one of the networks and the request came from the client within one
	// of the networks.  It must not be nil.
	PrivateSubnets netutil.SubnetSet

	// Cache is the configuration for the DNS results cache.  It must not be
	// nil.
	Cache *CacheConfig

	// Bootstrap describes bootstrapping DNS servers.  It must not be nil.
	Bootstrap *BootstrapConfig

	// Upstreams describes DNS upstream servers.  It must not be nil.
	Upstreams *UpstreamConfig

	// Fallbacks describes DNS fallback upstream servers.  It must not be nil.
	Fallbacks *FallbackConfig

	// ClientGetter is the function to get the client for a request.  It must
	// not be nil.
	ClientGetter ClientGetter

	// ListenAddrs is the list of served addresses.  It must contain at least a
	// single entry.
	ListenAddrs []netip.AddrPort
}
