package dnssvc

import (
	"fmt"
	"net/netip"

	"github.com/AdguardTeam/dnsproxy/proxy"
)

// upstreamConfigs is a set of client-specific upstream configurations.
type upstreamConfigs map[netip.Prefix]*proxy.UpstreamConfig

// clients creates a list of clients from confs.
func (confs upstreamConfigs) clients(cacheConf *CacheConfig) (clients []*client) {
	for cli, conf := range confs {
		clients = append(clients, &client{
			conf: proxy.NewCustomUpstreamConfig(
				conf,
				cacheConf.Enabled,
				cacheConf.ClientSize,
				false,
			),
			prefix: cli,
		})
	}

	return clients
}

// clientStorage stores clients and their upstream configurations.
type clientStorage struct {
	// clients is the actual list of existing clients.
	//
	// TODO(e.burkov):  Think of a way to make search more efficient.
	clients []*client
}

// newClientStorage creates a new storage of clients.
func newClientStorage(clients []*client) (cs *clientStorage) {
	return &clientStorage{
		clients: clients,
	}
}

// client stores the upstream configuration and the prefix for clients that
// should use it.
//
// TODO(e.burkov):  Think of a better name for this type.
type client struct {
	conf   *proxy.CustomUpstreamConfig
	prefix netip.Prefix
}

// find returns the client by its address or nil if no such clients exist.  The
// returned client is not a copy, so it must not be modified.
func (cs *clientStorage) find(addr netip.Addr) (c *client) {
	// TODO(e.burkov):  Handle overlapping prefixes.  Perhaps, choose the
	// narrowest.
	for _, cli := range cs.clients {
		if cli.prefix.Contains(addr) {
			return cli
		}
	}

	return nil
}

// close closes the storage and the upstream configurations of all its clients.
// It returns a slice of errors that occurred during the closing.  It must not
// be used concurrently with any existing client, i.e. any DNS processing must
// be stopped beforeward.
func (cs *clientStorage) close() (errs []error) {
	for _, c := range cs.clients {
		err := c.conf.Close()
		if err != nil {
			err = fmt.Errorf("closing upstreams for client %s: %w", c.prefix, err)

			errs = append(errs, err)
		}
	}

	return errs
}
