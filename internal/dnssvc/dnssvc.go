// Package dnssvc provides DNS handling functionality for AdGuardDNSClient.
package dnssvc

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"
	"slices"
	"time"

	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/service"
)

// DNSService is a service that provides DNS handling functionality.
type DNSService struct {
	// proxy forwards DNS requests.
	proxy *proxy.Proxy

	// bootstrapUpstreams is a list of upstreams to close on shutdown.
	bootstrapUpstreams []io.Closer
}

// New creates a new DNSService.  conf must not be nil.
func New(conf *Config) (svc *DNSService, err error) {
	// TODO(e.burkov):  Other protocols.
	udpListenAddrs := make([]*net.UDPAddr, 0, len(conf.ListenAddrs))
	tcpListenAddrs := make([]*net.TCPAddr, 0, len(conf.ListenAddrs))
	for _, addr := range conf.ListenAddrs {
		udpListenAddrs = append(udpListenAddrs, net.UDPAddrFromAddrPort(addr))
		tcpListenAddrs = append(tcpListenAddrs, net.TCPAddrFromAddrPort(addr))
	}

	boot, bootUps, err := newBootstrap(conf.Bootstrap)
	if err != nil {
		return nil, fmt.Errorf("creating bootstrap: %w", err)
	}

	ups, err := newUpstreams(conf.Upstreams.Addresses, conf.Upstreams.Timeout, boot)
	if err != nil {
		return nil, fmt.Errorf("creating upstreams: %w", err)
	}

	fall, err := newUpstreams(conf.Fallbacks.Addresses, conf.Fallbacks.Timeout, boot)
	if err != nil {
		return nil, fmt.Errorf("creating fallbacks: %w", err)
	}

	prx := &proxy.Proxy{
		Config: proxy.Config{
			UDPListenAddr: udpListenAddrs,
			TCPListenAddr: tcpListenAddrs,
			// TODO(e.burkov):  Create properly.
			UpstreamConfig: &proxy.UpstreamConfig{
				Upstreams: ups,
			},
			Fallbacks: &proxy.UpstreamConfig{
				Upstreams: fall,
			},
			// TODO(e.burkov):  Consider making configurable.
			TrustedProxies: netutil.SliceSubnetSet{
				netip.MustParsePrefix("0.0.0.0/0"),
				netip.MustParsePrefix("::/0"),
			},
		},
	}

	err = prx.Init()
	if err != nil {
		return nil, fmt.Errorf("initializing proxy: %w", err)
	}

	return &DNSService{
		proxy:              prx,
		bootstrapUpstreams: bootUps,
	}, nil
}

// type check
var _ service.Interface = (*DNSService)(nil)

// Start implements the [service.Interface] interface for *DNSService.
func (s *DNSService) Start(_ context.Context) (err error) {
	return s.proxy.Start()
}

// Shutdown implements the [service.Interface] interface for *DNSService.
func (s *DNSService) Shutdown(ctx context.Context) (err error) {
	errs := s.closeBootstraps()
	err = s.proxy.Stop()
	if err != nil {
		errs = append(errs, fmt.Errorf("stopping proxy: %w", err))
	}

	return errors.Join(errs...)
}

// closeBootstraps closes all bootstraps and returns all the errors joined.
func (s *DNSService) closeBootstraps() (errs []error) {
	for i, u := range s.bootstrapUpstreams {
		err := u.Close()
		if err != nil {
			errs = append(errs, fmt.Errorf("bootstrap at index %d: %w", i, err))
		}
	}

	return errs
}

// newBootstrap creates a new bootstrap resolver and a list of upstreams to
// close on shutdown.
func newBootstrap(conf *BootstrapConfig) (boot upstream.Resolver, closers []io.Closer, err error) {
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
			err = fmt.Errorf("creating bootstrap at index %d: %w", i, err)
			errs = append(errs, err)

			continue
		}

		resolvers = append(resolvers, upstream.NewCachingResolver(b))
		closers = append(closers, b.Upstream)
	}

	return slices.Clip(resolvers), slices.Clip(closers), errors.Join(errs...)
}

// newUpstreams creates a slice of upstreams from the given configuration.
func newUpstreams(
	addrs []string,
	timeout time.Duration,
	boot upstream.Resolver,
) (ups []upstream.Upstream, err error) {
	opts := &upstream.Options{
		Timeout:   timeout,
		Bootstrap: boot,
	}

	var errs []error
	for i, addr := range addrs {
		var u upstream.Upstream
		u, err = upstream.AddressToUpstream(addr, opts)
		if err != nil {
			errs = append(errs, fmt.Errorf("upstream at index %d: %w", i, err))

			continue
		}

		ups = append(ups, u)
	}

	return ups, errors.Join(errs...)
}
