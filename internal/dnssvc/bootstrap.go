package dnssvc

import (
	"fmt"
	"io"
	"net/netip"
	"time"

	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/AdguardTeam/golibs/errors"
)

// BootstrapConfig is the configuration for DNS bootstrap servers.
type BootstrapConfig struct {
	// Addresses is the list of servers.
	Addresses []netip.AddrPort

	// Timeout is the timeout for DNS requests.
	Timeout time.Duration
}

// newResolvers creates a new bootstrap resolver and a list of upstreams to
// close on shutdown.
func newResolvers(conf *BootstrapConfig) (boot upstream.Resolver, closers []io.Closer, err error) {
	defer func() { err = errors.Annotate(err, "creating bootstraps: %w") }()

	opts := &upstream.Options{
		Timeout: conf.Timeout,
	}

	resolvers := make(upstream.ConsequentResolver, 0, len(conf.Addresses))
	closers = make([]io.Closer, 0, len(conf.Addresses))

	var errs []error
	for i, addr := range conf.Addresses {
		var b *upstream.UpstreamResolver
		b, err = upstream.NewUpstreamResolver(addr.String(), opts)
		if err != nil {
			err = fmt.Errorf("address at index %d: %w", i, err)
			errs = append(errs, err)

			continue
		}

		resolvers = append(resolvers, upstream.NewCachingResolver(b))
		closers = append(closers, b.Upstream)
	}

	return resolvers, closers, errors.Join(errs...)
}
