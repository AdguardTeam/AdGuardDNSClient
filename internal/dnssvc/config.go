package dnssvc

import (
	"net/netip"
	"time"
)

// BootstrapConfig is the configuration for DNS bootstrap servers.
type BootstrapConfig struct {
	// Addresses is the list of servers.
	Addresses []netip.AddrPort

	// Timeout is the timeout for DNS requests.
	Timeout time.Duration
}

// UpstreamConfig is the configuration for DNS upstream servers.
type UpstreamConfig struct {
	// Addresses is the list of servers.
	Addresses []string

	// Timeout is the timeout for DNS requests.
	Timeout time.Duration
}

// FallbackConfig is the configuration for DNS fallback upstream servers.
type FallbackConfig struct {
	// Addresses is the list of servers.
	Addresses []string

	// Timeout is the timeout for DNS requests.
	Timeout time.Duration
}

// Config is the configuration for [DNSService].
//
// TODO(e.burkov):  Add cache.
//
// TODO(e.burkov):  Add contracts.
type Config struct {
	// Bootstrap describes bootstrapping DNS servers.
	Bootstrap *BootstrapConfig

	// Upstreams describes DNS upstream servers.
	Upstreams *UpstreamConfig

	// Fallbacks describes DNS fallback upstream servers.
	Fallbacks *FallbackConfig

	// ListenAddrs is the list of served addresses.
	ListenAddrs []netip.AddrPort
}
