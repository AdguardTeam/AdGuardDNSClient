// Package dnssvc provides DNS handling functionality for AdGuardDNSClient.
package dnssvc

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"

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

	boot, bootUps, err := conf.Bootstrap.toProxyResolvers()
	if err != nil {
		return nil, fmt.Errorf("creating bootstrap: %w", err)
	}

	ups, err := conf.Upstreams.toProxyUpstreamConfig(&upstream.Options{
		Timeout:   conf.Upstreams.Timeout,
		Bootstrap: boot,
	})
	if err != nil {
		return nil, fmt.Errorf("creating upstreams: %w", err)
	}

	fallbacks, err := proxy.ParseUpstreamsConfig(conf.Fallbacks.Addresses, &upstream.Options{
		Timeout:   conf.Fallbacks.Timeout,
		Bootstrap: boot,
	})
	if err != nil {
		return nil, fmt.Errorf("creating fallbacks: %w", err)
	}

	prx, err := proxy.New(&proxy.Config{
		UDPListenAddr:  udpListenAddrs,
		TCPListenAddr:  tcpListenAddrs,
		UpstreamConfig: ups,
		Fallbacks:      fallbacks,
		// TODO(e.burkov):  Consider making configurable.
		TrustedProxies: netutil.SliceSubnetSet{
			netip.MustParsePrefix("0.0.0.0/0"),
			netip.MustParsePrefix("::/0"),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating proxy: %w", err)
	}

	return &DNSService{
		proxy:              prx,
		bootstrapUpstreams: bootUps,
	}, nil
}

// type check
var _ service.Interface = (*DNSService)(nil)

// Start implements the [service.Interface] interface for *DNSService.
func (s *DNSService) Start(ctx context.Context) (err error) {
	return s.proxy.Start(ctx)
}

// Shutdown implements the [service.Interface] interface for *DNSService.
func (s *DNSService) Shutdown(ctx context.Context) (err error) {
	var errs []error

	err = s.proxy.Shutdown(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("stopping proxy: %w", err))
	}

	// Close bootstraps after upstreams to ensure those aren't used anymore.
	errs = append(errs, s.closeBootstraps()...)

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
