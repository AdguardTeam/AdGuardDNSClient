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
		Enabled:    c.Enabled,
		Size:       c.Size,
		ClientSize: c.ClientSize,
	}
}

// type check
var _ validator = (*cacheConfig)(nil)

// validate implements the [validator] interface for *cacheConfig.
func (c *cacheConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "cache: %w") }()

	if c == nil {
		return errNoValue
	} else if !c.Enabled {
		// Don't validate cache settings if it's disabled.
		return nil
	}

	var errs []error

	// TODO(e.burkov):  Remove [math.MaxInt] constraint when [datasize.ByteSize]
	// is supported by proxy.

	if c.Size == 0 {
		err = fmt.Errorf("got size %s: %w", c.Size, errMustBePositive)
		errs = append(errs, err)
	} else if c.Size > math.MaxInt {
		err = fmt.Errorf("got size %s: must be less or equal to %d", c.Size, math.MaxInt)
		errs = append(errs, err)
	}

	if c.ClientSize == 0 {
		err = fmt.Errorf("got client_size %s: %w", c.ClientSize, errMustBePositive)
		errs = append(errs, err)
	} else if c.ClientSize > math.MaxInt {
		err = fmt.Errorf("got size %s: must be less or equal to %d", c.ClientSize, math.MaxInt)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
