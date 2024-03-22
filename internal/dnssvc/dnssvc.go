// Package dnssvc provides DNS handling functionality for AdGuardDNSClient.
package dnssvc

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"

	"github.com/AdguardTeam/dnsproxy/proxy"
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
	prxConf, bootUps, err := newProxyConfig(conf)
	if err != nil {
		return nil, fmt.Errorf("creating proxy configuration: %w", err)
	}

	prx, err := proxy.New(prxConf)
	if err != nil {
		return nil, fmt.Errorf("creating proxy: %w", err)
	}

	return &DNSService{
		proxy:              prx,
		bootstrapUpstreams: bootUps,
	}, nil
}

// newProxyConfig creates a new [proxy.Config] from conf.  conf must not be nil.
func newProxyConfig(conf *Config) (prxConf *proxy.Config, bootUps []io.Closer, err error) {
	boot, bootUps, err := newResolvers(conf.Bootstrap)
	if err != nil {
		// Don't wrap the error, because it's informative enough as is.
		return nil, nil, err
	}

	ups, err := newUpstreams(conf.Upstreams, boot)
	if err != nil {
		// Don't wrap the error, because it's informative enough as is.
		return nil, bootUps, err
	}

	falls, err := newFallbacks(conf.Fallbacks, boot)
	if err != nil {
		// Don't wrap the error, because it's informative enough as is.
		return nil, bootUps, err
	}

	udp, tcp := newListenAddrs(conf.ListenAddrs)
	// TODO(e.burkov):  Consider making configurable.
	trusted := netutil.SliceSubnetSet{
		netip.PrefixFrom(netip.IPv4Unspecified(), 0),
		netip.PrefixFrom(netip.IPv6Unspecified(), 0),
	}

	return &proxy.Config{
		UDPListenAddr:  udp,
		TCPListenAddr:  tcp,
		UpstreamConfig: ups,
		Fallbacks:      falls,
		TrustedProxies: trusted,
	}, bootUps, nil
}

// newListenAddrs creates a new list of UDP and TCP addresses from addrs.
//
// TODO(e.burkov):  Support other protos.
func newListenAddrs(addrs []netip.AddrPort) (udp []*net.UDPAddr, tcp []*net.TCPAddr) {
	udp = make([]*net.UDPAddr, 0, len(addrs))
	tcp = make([]*net.TCPAddr, 0, len(addrs))
	for _, addr := range addrs {
		udp = append(udp, net.UDPAddrFromAddrPort(addr))
		tcp = append(tcp, net.TCPAddrFromAddrPort(addr))
	}

	return udp, tcp
}

// type check
var _ service.Interface = (*DNSService)(nil)

// Start implements the [service.Interface] interface for *DNSService.
func (svc *DNSService) Start(ctx context.Context) (err error) {
	return svc.proxy.Start(ctx)
}

// Shutdown implements the [service.Interface] interface for *DNSService.
func (svc *DNSService) Shutdown(ctx context.Context) (err error) {
	var errs []error

	err = svc.proxy.Shutdown(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("stopping proxy: %w", err))
	}

	errs = append(errs, svc.closeBootstraps()...)

	return errors.Join(errs...)
}

// closeBootstraps closes all bootstraps and returns all the errors joined.
func (svc *DNSService) closeBootstraps() (errs []error) {
	for i, u := range svc.bootstrapUpstreams {
		err := u.Close()
		if err != nil {
			err = fmt.Errorf("closing bootstrap at index %d: %w", i, err)

			errs = append(errs, err)
		}
	}

	return errs
}
