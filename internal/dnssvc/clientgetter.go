package dnssvc

import (
	"net/netip"

	"github.com/AdguardTeam/dnsproxy/proxy"
)

// ClientGetter retrieves the client's address from the DNS context.
type ClientGetter interface {
	// Address returns the client's address.
	Address(dctx *proxy.DNSContext) (addr netip.AddrPort)
}

// DefaultClientGetter is a default implementation of ClientGetter.
type DefaultClientGetter struct{}

// type check
var _ ClientGetter = DefaultClientGetter{}

// Address implements the ClientGetter interface for defaultClientGetter.
func (DefaultClientGetter) Address(dctx *proxy.DNSContext) (addr netip.AddrPort) {
	return dctx.Addr
}
