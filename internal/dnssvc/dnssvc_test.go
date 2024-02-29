package dnssvc_test

import (
	"context"
	"net/netip"
	"net/url"
	"testing"
	"time"

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
	req := (&dns.Msg{}).SetQuestion("example.com.", dns.TypeA)
	resp := (&dns.Msg{}).SetReply(req)
	upsHandler := func(w dns.ResponseWriter, _ *dns.Msg) {
		require.NoError(testutil.PanicT{}, w.WriteMsg(resp))
	}

	upsURL := startLocalhostUpstream(t, dns.HandlerFunc(upsHandler))

	svc, err := dnssvc.New(&dnssvc.Config{
		Bootstrap: &dnssvc.BootstrapConfig{},
		Upstreams: &dnssvc.UpstreamConfig{
			Addresses: []string{upsURL.String()},
		},
		Fallbacks: &dnssvc.FallbackConfig{
			Addresses: []string{upsURL.String()},
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
	tcpAddr := svc.Addr(proxy.ProtoTCP)

	received, _, err := cli.Exchange(req, tcpAddr.String())
	require.NoError(t, err)

	assert.Equal(t, resp, received)
}
