package dnssvc

import (
	"net/netip"
	"testing"

	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/stretchr/testify/assert"
)

func TestClientStorage(t *testing.T) {
	t.Parallel()

	cli1Addr1 := netip.MustParseAddr("1.2.3.4")
	cli1Pref := netip.PrefixFrom(cli1Addr1, 24)
	cli1Addr2 := cli1Addr1.Next()

	cli2Addr1 := netip.MustParseAddr("4.3.2.1")
	cli2Pref := netip.PrefixFrom(cli2Addr1, 32)
	absentAddr := cli2Addr1.Next()

	cli1 := &client{
		prefix: cli1Pref,
		conf:   &proxy.CustomUpstreamConfig{},
	}
	cli2 := &client{
		prefix: cli2Pref,
		conf:   &proxy.CustomUpstreamConfig{},
	}

	// search is a case of searching through a particular clients set.
	type search struct {
		addr netip.Addr
		want *client
	}

	testCases := []struct {
		name     string
		clients  []*client
		searches []search
	}{{
		name:    "empty",
		clients: nil,
		searches: []search{{
			addr: cli1Addr1,
			want: nil,
		}, {
			addr: cli1Addr2,
			want: nil,
		}, {
			addr: cli2Addr1,
			want: nil,
		}, {
			addr: absentAddr,
			want: nil,
		}},
	}, {
		name: "single",
		clients: []*client{
			cli1,
		},
		searches: []search{{
			addr: cli1Addr1,
			want: cli1,
		}, {
			addr: cli1Addr2,
			want: cli1,
		}, {
			addr: cli2Addr1,
			want: nil,
		}, {
			addr: absentAddr,
			want: nil,
		}},
	}, {
		name: "multiple",
		clients: []*client{
			cli1,
			cli2,
		},
		searches: []search{{
			addr: cli1Addr1,
			want: cli1,
		}, {
			addr: cli1Addr2,
			want: cli1,
		}, {
			addr: cli2Addr1,
			want: cli2,
		}, {
			addr: absentAddr,
			want: nil,
		}},
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs := newClientStorage(tc.clients)
			testutil.CleanupAndRequireSuccess(t, func() (err error) {
				return errors.Join(cs.close()...)
			})

			for _, sc := range tc.searches {
				t.Run(sc.addr.String(), func(t *testing.T) {
					t.Parallel()

					assert.Same(t, sc.want, cs.find(sc.addr))
				})
			}
		})
	}
}
