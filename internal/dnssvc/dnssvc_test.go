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

// TODO(e.burkov):  Add bootstrap.
func TestDNSService(t *testing.T) {
	t.Parallel()

	const (
		testDomain    = "example.com"
		testSubdomain = "test.example.com"
	)

	commonReq := (&dns.Msg{}).SetQuestion(dns.Fqdn(testDomain), dns.TypeA)
	commonReq.Id = 1
	commonResp := (&dns.Msg{}).SetReply(commonReq)

	subdomainReq := (&dns.Msg{}).SetQuestion(dns.Fqdn(testSubdomain), dns.TypeA)
	subdomainReq.Id = 2
	subdomainResp := (&dns.Msg{}).SetReply(subdomainReq)

	pt := testutil.PanicT{}
	commonUps := func(w dns.ResponseWriter, _ *dns.Msg) {
		require.NoError(pt, w.WriteMsg(commonResp))
	}
	subdomainUps := func(w dns.ResponseWriter, _ *dns.Msg) {
		require.NoError(pt, w.WriteMsg(subdomainResp))
	}

	commonURL := startLocalhostUpstream(t, dns.HandlerFunc(commonUps)).String()
	subdomainURL := startLocalhostUpstream(t, dns.HandlerFunc(subdomainUps)).String()

	svc, err := dnssvc.New(&dnssvc.Config{
		Bootstrap: &dnssvc.BootstrapConfig{},
		Upstreams: &dnssvc.UpstreamConfig{
			Groups: []*dnssvc.UpstreamGroupConfig{{
				Name:    agdc.UpstreamGroupNameDefault,
				Address: commonURL,
			}, {
				Name:    "domain-group",
				Address: subdomainURL,
				Match: []dnssvc.MatchCriteria{{
					QuestionDomain: testSubdomain,
				}},
			}},
			Timeout: testTimeout,
		},
		Fallbacks: &dnssvc.FallbackConfig{
			Addresses: []string{commonURL},
			Timeout:   testTimeout,
		},
		ListenAddrs: []netip.AddrPort{
			netip.AddrPortFrom(netutil.IPv4Localhost(), 0),
		},
	})
	require.NoError(t, err)

	ctx := context.Background()
	err = svc.Start(ctx)
	require.NoError(t, err)
	testutil.CleanupAndRequireSuccess(t, func() (err error) { return svc.Shutdown(ctx) })

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
