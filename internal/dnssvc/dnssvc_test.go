package dnssvc_test

import (
	"context"
	"net/netip"
	"net/url"
	"testing"
	"time"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdc"
	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testTimeout is a timeout for tests.
//
// TODO(e.burkov):  Move into agdctest.
const testTimeout = 1 * time.Second

// startLocalhostUpstream is a test helper that starts a DNS server on
// localhost.
//
// TODO(e.burkov):  Move into agdctest or even to golibs.
func startLocalhostUpstream(t *testing.T, h dns.Handler) (addr *url.URL) {
	t.Helper()

	startCh := make(chan netip.AddrPort)
	errCh := make(chan error)

	srv := &dns.Server{
		Addr:         netip.AddrPortFrom(netutil.IPv4Localhost(), 0).String(),
		Net:          string(proxy.ProtoTCP),
		Handler:      h,
		ReadTimeout:  testTimeout,
		WriteTimeout: testTimeout,
	}
	srv.NotifyStartedFunc = func() {
		addrPort := srv.Listener.Addr()
		startCh <- netutil.NetAddrToAddrPort(addrPort)
	}

	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case addrPort := <-startCh:
		addr = &url.URL{
			Scheme: string(proxy.ProtoTCP),
			Host:   addrPort.String(),
		}

		testutil.CleanupAndRequireSuccess(t, func() (err error) { return <-errCh })
		testutil.CleanupAndRequireSuccess(t, srv.Shutdown)
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(testTimeout):
		require.FailNow(t, "timeout exceeded")
	}

	return addr
}

// testClientGetter is a mock implementation of [dnssvc.ClientGetter] for tests.
type testClientGetter struct {
	OnAddress func(dctx *proxy.DNSContext) (addr netip.AddrPort)
}

// type check
var _ dnssvc.ClientGetter = (*testClientGetter)(nil)

// Address implements the [dnssvc.ClientGetter] interface for *testClientGetter.
func (cg *testClientGetter) Address(dctx *proxy.DNSContext) (addr netip.AddrPort) {
	return cg.OnAddress(dctx)
}

// TODO(e.burkov):  Add bootstrap.
func TestDNSService(t *testing.T) {
	t.Parallel()

	// Declare domain names.

	const (
		testDomain    = "example.com"
		testSubdomain = "test.example.com"
	)

	privateNets := netutil.SubnetSetFunc(netutil.IsLocallyServed)

	privateAddr := netip.MustParseAddr("192.168.1.1")
	require.True(t, privateNets.Contains(privateAddr))

	privateARPADomain, err := netutil.IPToReversedAddr(privateAddr.AsSlice())
	require.NoError(t, err)

	// Create requests and responses.

	commonReq := (&dns.Msg{}).SetQuestion(dns.Fqdn(testDomain), dns.TypeA)
	commonReq.Id = 1
	commonResp := (&dns.Msg{}).SetReply(commonReq)

	subdomainReq := (&dns.Msg{}).SetQuestion(dns.Fqdn(testSubdomain), dns.TypeA)
	subdomainReq.Id = 2
	subdomainResp := (&dns.Msg{}).SetReply(subdomainReq)

	cliSpecReq := (&dns.Msg{}).SetQuestion(dns.Fqdn(testDomain), dns.TypeA)
	cliSpecReq.Id = 3
	cliSpecResp := (&dns.Msg{}).SetReply(cliSpecReq)

	subdomainCliSpecReq := (&dns.Msg{}).SetQuestion(dns.Fqdn(testSubdomain), dns.TypeA)
	subdomainCliSpecReq.Id = 4
	subdomainCliSpecResp := (&dns.Msg{}).SetReply(subdomainCliSpecReq)

	privateReq := (&dns.Msg{}).SetQuestion(dns.Fqdn(privateARPADomain), dns.TypePTR)
	privateReq.Id = 5
	privateResp := (&dns.Msg{}).SetReply(privateReq)

	forbiddenReq := (&dns.Msg{}).SetQuestion(dns.Fqdn(privateARPADomain), dns.TypePTR)
	forbiddenReq.Id = 6
	forbiddenResp := (&dns.Msg{}).SetRcode(forbiddenReq, dns.RcodeNameError)
	forbiddenResp.RecursionAvailable = true

	// Create upstreams.

	pt := testutil.PanicT{}
	commonUps := func(w dns.ResponseWriter, _ *dns.Msg) {
		require.NoError(pt, w.WriteMsg(commonResp))
	}
	subdomainUps := func(w dns.ResponseWriter, _ *dns.Msg) {
		require.NoError(pt, w.WriteMsg(subdomainResp))
	}
	cliSpecUps := func(w dns.ResponseWriter, _ *dns.Msg) {
		require.NoError(pt, w.WriteMsg(cliSpecResp))
	}
	subdomainCliSpecUps := func(w dns.ResponseWriter, _ *dns.Msg) {
		require.NoError(pt, w.WriteMsg(subdomainCliSpecResp))
	}
	privateUps := func(w dns.ResponseWriter, _ *dns.Msg) {
		require.NoError(pt, w.WriteMsg(privateResp))
	}

	commonURL := startLocalhostUpstream(t, dns.HandlerFunc(commonUps)).String()
	subdomainURL := startLocalhostUpstream(t, dns.HandlerFunc(subdomainUps)).String()
	cliSpecURL := startLocalhostUpstream(t, dns.HandlerFunc(cliSpecUps)).String()
	subdomainCliSpecURL := startLocalhostUpstream(t, dns.HandlerFunc(subdomainCliSpecUps)).String()
	privateURL := startLocalhostUpstream(t, dns.HandlerFunc(privateUps)).String()

	// Prepare clients.

	cli1Addr := netip.MustParseAddr("1.2.3.4")
	cli2Addr := netip.MustParseAddr("4.3.2.1")
	cli2Pref := netip.PrefixFrom(cli2Addr, cli2Addr.BitLen())

	privateCli := netip.MustParseAddr("192.168.1.2")
	require.True(t, privateNets.Contains(privateCli))

	externalCli := netip.MustParseAddr("123.123.123.123")
	require.False(t, privateNets.Contains(externalCli))

	cliGetter := &testClientGetter{
		OnAddress: func(dctx *proxy.DNSContext) (addr netip.AddrPort) {
			switch dctx.Req.Id {
			case commonReq.Id, subdomainReq.Id:
				return netip.AddrPortFrom(cli1Addr, 1)
			case cliSpecReq.Id, subdomainCliSpecReq.Id:
				return netip.AddrPortFrom(cli2Addr, 1)
			case privateReq.Id:
				return netip.AddrPortFrom(privateCli, 1)
			case forbiddenReq.Id:
				return netip.AddrPortFrom(externalCli, 1)
			default:
				panic("unexpected request")
			}
		},
	}

	// Create and start the service.

	svc, err := dnssvc.New(&dnssvc.Config{
		Logger:         slogutil.NewDiscardLogger(),
		PrivateSubnets: privateNets,
		Bootstrap:      &dnssvc.BootstrapConfig{},
		Cache: &dnssvc.CacheConfig{
			Enabled: false,
		},
		Upstreams: &dnssvc.UpstreamConfig{
			Groups: []*dnssvc.UpstreamGroupConfig{{
				Name:    agdc.UpstreamGroupNameDefault,
				Address: commonURL,
			}, {
				Name:    agdc.UpstreamGroupNamePrivate,
				Address: privateURL,
			}, {
				Name:    "domain-group",
				Address: subdomainURL,
				Match: []dnssvc.MatchCriteria{{
					QuestionDomain: testSubdomain,
				}},
			}, {
				Name:    "client-group",
				Address: cliSpecURL,
				Match: []dnssvc.MatchCriteria{{
					Client: cli2Pref,
				}},
			}, {
				Name:    "domain-client-group",
				Address: subdomainCliSpecURL,
				Match: []dnssvc.MatchCriteria{{
					Client:         cli2Pref,
					QuestionDomain: testSubdomain,
				}},
			}},
			Timeout: testTimeout,
		},
		Fallbacks: &dnssvc.FallbackConfig{
			Addresses: []string{
				commonURL,
			},
			Timeout: testTimeout,
		},
		ClientGetter: cliGetter,
		ListenAddrs: []netip.AddrPort{
			netip.AddrPortFrom(netutil.IPv4Localhost(), 0),
		},
	})
	require.NoError(t, err)

	ctx := context.Background()
	err = svc.Start(ctx)
	require.NoError(t, err)
	testutil.CleanupAndRequireSuccess(t, func() (err error) { return svc.Shutdown(ctx) })

	// Test.

	tcpAddr := svc.Addr(proxy.ProtoTCP).String()

	cli := &dns.Client{
		Net:     string(proxy.ProtoTCP),
		Timeout: testTimeout,
	}

	testCases := []struct {
		req      *dns.Msg
		wantResp *dns.Msg
		name     string
	}{{
		req:      commonReq,
		wantResp: commonResp,
		name:     "success",
	}, {
		req:      subdomainReq,
		wantResp: subdomainResp,
		name:     "domain_match_success",
	}, {
		req:      cliSpecReq,
		wantResp: cliSpecResp,
		name:     "client_match_success",
	}, {
		req:      subdomainCliSpecReq,
		wantResp: subdomainCliSpecResp,
		name:     "domain_client_match_success",
	}, {
		req:      privateReq,
		wantResp: privateResp,
		name:     "private_success",
	}, {
		req:      forbiddenReq,
		wantResp: forbiddenResp,
		name:     "private_forbidden",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			received, _, excErr := cli.Exchange(tc.req, tcpAddr)
			require.NoError(t, excErr)
			assert.Equal(t, tc.wantResp, received)
		})
	}
}
