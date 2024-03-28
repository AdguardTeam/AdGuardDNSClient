package dnssvc

import (
	"net/netip"
)

// Config is the configuration for [DNSService].
//
// TODO(e.burkov):  Add cache.
type Config struct {
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
