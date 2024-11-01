package dnssvc

// CacheConfig is the configuration for the DNS results cache.
type CacheConfig struct {
	// Enabled specifies if the cache should be used.
	Enabled bool

	// Size is the maximum size of the common cache.
	//
	// TODO(e.burkov):  Make it a [datasize.ByteSize] when dnsproxy uses it.
	Size int

	// ClientSize is the maximum size of each per-client cache.
	//
	// TODO(e.burkov):  Make it a [datasize.ByteSize] when dnsproxy uses it.
	ClientSize int
}

// TODO(e.burkov):  Add tests.
