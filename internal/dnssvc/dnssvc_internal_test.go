package dnssvc

import (
	"net"

	"github.com/AdguardTeam/dnsproxy/proxy"
)

// Addr returns the address of the service for the given protocol.  This is only
// needed for testing.
func (svc *DNSService) Addr(proto proxy.Proto) (addr net.Addr) {
	return svc.proxy.Addr(proto)
}
