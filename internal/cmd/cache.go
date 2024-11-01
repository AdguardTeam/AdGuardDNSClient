package cmd

import (
	"fmt"
	"math"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/c2h5oh/datasize"
)

// cacheConfig is the configuration for the DNS results cache.
type cacheConfig struct {
	// Enabled specifies if the cache should be used.
	Enabled bool `yaml:"enabled"`

	// Size is the maximum size of the cache.
	Size datasize.ByteSize `yaml:"size"`

	// ClientSize is the maximum size of the cache per client.
	ClientSize datasize.ByteSize `yaml:"client_size"`
}

// toInternal converts the cache configuration to the internal representation.
// c must be valid.
func (c *cacheConfig) toInternal() (conf *dnssvc.CacheConfig) {
	return &dnssvc.CacheConfig{
		Enabled: c.Enabled,
		// #nosec G115 -- The value is validated to not exceed [math.MaxInt].
		Size: int(c.Size),
		// #nosec G115 -- The value is validated to not exceed [math.MaxInt].
		ClientSize: int(c.ClientSize),
	}
}

// type check
var _ validator = (*cacheConfig)(nil)

// validate implements the [validator] interface for *cacheConfig.
func (c *cacheConfig) validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	} else if !c.Enabled {
		// Don't validate cache settings if it's disabled.
		return nil
	}

	var errs []error

	// TODO(e.burkov):  Remove [math.MaxInt] constraint when [datasize.ByteSize]
	// is supported by proxy.

	if c.Size == 0 || c.Size > math.MaxInt {
		err = fmt.Errorf(
			"size: %w: must be positive and less than %d; got %d",
			errors.ErrOutOfRange,
			math.MaxInt,
			c.Size,
		)
		errs = append(errs, err)
	}

	if c.Size == 0 || c.Size > math.MaxInt {
		err = fmt.Errorf(
			"client_size: %w: must be positive and less than %d; got %d",
			errors.ErrOutOfRange,
			math.MaxInt,
			c.Size,
		)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
