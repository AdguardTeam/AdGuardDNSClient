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
func startLocalhostUpstream(t *testing.T, h dns.Handler) (addr *url.URL) {
	t.Helper()

	startCh := make(chan netip.AddrPort)
	defer close(startCh)
	errCh := make(chan error)

	srv := &dns.Server{
		Addr:         "127.0.0.1:0",
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

func TestDNSService(t *testing.T) {
	t.Parallel()

	req1 := (&dns.Msg{}).SetQuestion("example.com.", dns.TypeA)
	resp1 := (&dns.Msg{}).SetReply(req1)

	req2 := (&dns.Msg{}).SetQuestion("test.example.com.", dns.TypeA)
	resp2 := (&dns.Msg{}).SetReply(req2)

	upsHdlr1 := func(w dns.ResponseWriter, req *dns.Msg) {
		require.NoError(testutil.PanicT{}, w.WriteMsg(resp1))
	}

	upsHdlr2 := func(w dns.ResponseWriter, req *dns.Msg) {
		require.NoError(testutil.PanicT{}, w.WriteMsg(resp2))
	}

	ups1URL := startLocalhostUpstream(t, dns.HandlerFunc(upsHdlr1)).String()
	ups2URL := startLocalhostUpstream(t, dns.HandlerFunc(upsHdlr2)).String()

	svc, err := dnssvc.New(&dnssvc.Config{
		Bootstrap: &dnssvc.BootstrapConfig{},
		Upstreams: &dnssvc.UpstreamConfig{
			Groups: []*dnssvc.UpstreamGroupConfig{{
				Name:    agdc.UpstreamGroupNameDefault,
				Address: ups1URL,
				Match:   nil,
			}, {
				Name:    "test_group_name",
				Address: ups2URL,
				Match: []dnssvc.MatchCriteria{{
					QuestionDomain: "test.example.com",
				}},
			}},
			Timeout: testTimeout,
		},
		Fallbacks: &dnssvc.FallbackConfig{
			Addresses: []string{ups1URL},
			Timeout:   testTimeout,
		},
		ListenAddrs: []netip.AddrPort{
			netip.MustParseAddrPort("127.0.0.1:0"),
		},
	})
	require.NoError(t, err)

	ctx := context.Background()
	err = svc.Start(ctx)
	require.NoError(t, err)
	testutil.CleanupAndRequireSuccess(t, func() (err error) { return svc.Shutdown(ctx) })

	cli := &dns.Client{
		Net:     string(proxy.ProtoTCP),
		Timeout: testTimeout,
	}
	tcpAddr := svc.Addr(proxy.ProtoTCP).String()

	testCases := []struct {
		name       string
		req        *dns.Msg
		wantResp   *dns.Msg
		wantErrMsg string
	}{{
		name:     "success",
		req:      req1,
		wantResp: resp1,
	}, {
		name:     "domain_match",
		req:      req2,
		wantResp: resp2,
	}}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			received, _, excErr := cli.Exchange(tc.req, tcpAddr)
			testutil.AssertErrorMsg(t, tc.wantErrMsg, excErr)
			assert.Equal(t, tc.wantResp, received)
		})
	}
}
