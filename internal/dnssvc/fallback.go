package dnssvc

import (
	"fmt"
	"time"

	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/AdguardTeam/dnsproxy/upstream"
)

// FallbackConfig is the configuration for DNS fallback upstream servers.
type FallbackConfig struct {
	// Addresses is the list of servers.
	Addresses []string

	// Timeout is the timeout for DNS requests.  Zero value disables the
	// timeout.
	Timeout time.Duration
}

// newFallbacks creates a new fallback upstream configuration from conf using
// boot.
func newFallbacks(
	conf *FallbackConfig,
	boot upstream.Resolver,
) (fallbacks *proxy.UpstreamConfig, err error) {
	opts := &upstream.Options{
		Timeout:   conf.Timeout,
		Bootstrap: boot,
	}

	fallbacks, err = proxy.ParseUpstreamsConfig(conf.Addresses, opts)
	if err != nil {
		return nil, fmt.Errorf("creating fallbacks: %w", err)
	}

	return fallbacks, nil
}
