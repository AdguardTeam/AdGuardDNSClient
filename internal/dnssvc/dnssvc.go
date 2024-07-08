// Package dnssvc provides DNS handling functionality for AdGuardDNSClient.
package dnssvc

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"

	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/service"
)

// DNSService is a service that provides DNS handling functionality.
type DNSService struct {
	// logger is used as the base logger for the DNS service.
	logger *slog.Logger

	// proxy forwards DNS requests.
	proxy *proxy.Proxy

	// clients stores upstream configurations associated with clients'
	// addresses.
	clients *clientStorage

	// clientGetter is used to get the client's address from the request's
	// context.  It's only used for testing.
	//
	// TODO(e.burkov):  Put custom client's address to proxy context, when it
	// start supporting the [context.Context].  Then get rid of this interface.
	clientGetter ClientGetter

	// bootstrapUpstreams is a list of upstreams to close on shutdown.
	bootstrapUpstreams []io.Closer
}

// New creates a new DNSService.  conf must not be nil.
func New(conf *Config) (svc *DNSService, err error) {
	boot, bootUps, err := newResolvers(conf.Bootstrap)
	if err != nil {
		// Don't wrap the error, because it's informative enough as is.
		return nil, err
	}

	prxConf, clients, err := newProxyConfig(conf, boot)
	if err != nil {
		// Don't wrap the error, because it's informative enough as is.
		return nil, err
	}

	svc = &DNSService{
		logger:       conf.Logger.With(slogutil.KeyPrefix, "dnssvc"),
		clientGetter: conf.ClientGetter,
		clients:      newClientStorage(clients),
	}
	prxConf.BeforeRequestHandler = svc
	prxConf.RequestHandler = svc.handleRequest

	prx, err := proxy.New(prxConf)
	if err != nil {
		return nil, fmt.Errorf("creating proxy: %w", err)
	}

	svc.proxy = prx
	svc.bootstrapUpstreams = bootUps

	return svc, nil
}

// newProxyConfig creates a new [proxy.Config] from conf using boot for all
// upstream configurations.  It returns a ready-to-use configuration and clients
// with their specific upstream configurations.
func newProxyConfig(
	conf *Config,
	boot upstream.Resolver,
) (prxConf *proxy.Config, clients []*client, err error) {
	defer func() { err = errors.Annotate(err, "creating proxy configuration: %w") }()

	ups, private, err := newUpstreams(conf.Upstreams, boot)
	if err != nil {
		// Don't wrap the error, because it's informative enough as is.
		return nil, nil, err
	}

	falls, err := newFallbacks(conf.Fallbacks, boot)
	if err != nil {
		// Don't wrap the error, because it's informative enough as is.
		return nil, nil, err
	}

	// Use the upstream configuration with no client specification as the
	// general one.  Also remove it from the map, to build the clients list.
	general := ups[netip.Prefix{}]
	delete(ups, netip.Prefix{})

	udp, tcp := newListenAddrs(conf.ListenAddrs)
	// TODO(e.burkov):  Consider making configurable.
	trusted := netutil.SliceSubnetSet{
		netip.PrefixFrom(netip.IPv4Unspecified(), 0),
		netip.PrefixFrom(netip.IPv6Unspecified(), 0),
	}

	return &proxy.Config{
		Logger:                    conf.Logger.With(slogutil.KeyPrefix, "dnsproxy"),
		UpstreamMode:              proxy.UpstreamModeLoadBalance,
		UDPListenAddr:             udp,
		TCPListenAddr:             tcp,
		UpstreamConfig:            general,
		PrivateRDNSUpstreamConfig: private,
		PrivateSubnets:            conf.PrivateSubnets,
		UsePrivateRDNS:            private != nil,
		Fallbacks:                 falls,
		TrustedProxies:            trusted,
		CacheSizeBytes:            int(conf.Cache.Size),
		CacheEnabled:              conf.Cache.Enabled,
	}, ups.clients(conf.Cache), nil
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
	svc.logger.DebugContext(ctx, "starting")

	return svc.proxy.Start(ctx)
}

// Shutdown implements the [service.Interface] interface for *DNSService.
func (svc *DNSService) Shutdown(ctx context.Context) (err error) {
	svc.logger.DebugContext(ctx, "shutting down")

	var errs []error
	err = svc.proxy.Shutdown(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("stopping proxy: %w", err))
	}

	errs = append(errs, svc.clients.close()...)
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

// type check
var _ proxy.BeforeRequestHandler = (*DNSService)(nil)

// HandleBefore implements the [proxy.BeforeRequestHandler] interface for
// *DNSService.
func (svc *DNSService) HandleBefore(p *proxy.Proxy, dctx *proxy.DNSContext) (err error) {
	// This is used to substitute the client's address in tests.
	dctx.Addr = svc.clientGetter.Address(dctx)

	// Check the address privateness because proxy does it before substitution.
	// See TODO on [DNSService.clientGetter].
	dctx.IsPrivateClient = svc.proxy.PrivateSubnets.Contains(dctx.Addr.Addr())

	return nil
}

// handleRequest is a [proxy.RequestHandler].
func (svc *DNSService) handleRequest(p *proxy.Proxy, dctx *proxy.DNSContext) (err error) {
	if dctx.RequestedPrivateRDNS != (netip.Prefix{}) {
		// Don't match client for private PTR request.
		return p.Resolve(dctx)
	}

	c := svc.clients.find(dctx.Addr.Addr())
	if c != nil {
		dctx.CustomUpstreamConfig = c.conf
	}

	return p.Resolve(dctx)
}
