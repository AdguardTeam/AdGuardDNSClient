package dnssvc

import (
	"log/slog"
	"net/netip"

	"github.com/AdguardTeam/golibs/netutil"
)

// Config is the configuration for [DNSService].  All fields must not be empty.
type Config struct {
	// Logger is used as the base logger for the DNS service.
	Logger *slog.Logger

	// PrivateSubnets is the set of IP networks considered private.  The PTR
	// requests for ARPA domains considered private if the domain contains an IP
	// from one of the networks and the request came from the client within one
	// of the networks.
	PrivateSubnets netutil.SubnetSet

	// Cache is the configuration for the DNS results cache.
	Cache *CacheConfig

	// Bootstrap describes bootstrapping DNS servers.
	Bootstrap *BootstrapConfig

	// Upstreams describes DNS upstream servers.
	Upstreams *UpstreamConfig

	// Fallbacks describes DNS fallback upstream servers.
	Fallbacks *FallbackConfig

	// ClientGetter is the function to get the client for a request.
	ClientGetter ClientGetter

	// ListenAddrs is the list of served addresses.  It must contain at least
	// one entry.
	ListenAddrs []netip.AddrPort
}
