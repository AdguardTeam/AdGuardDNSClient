package cmd

import (
	"fmt"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/c2h5oh/datasize"
)

// cacheConfig is the configuration for the DNS results cache.
type cacheConfig struct {
	// Size is the maximum size of the cache.
	Size datasize.ByteSize `yaml:"size"`

	// ClientSize is the maximum size of the cache per client.
	ClientSize datasize.ByteSize `yaml:"client_size"`

	// Enabled specifies if the cache should be used.
	Enabled bool `yaml:"enabled"`
}

// type check
var _ validator = (*cacheConfig)(nil)

// validate implements the [validator] interface for *cacheConfig.
func (c *cacheConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "cache: %w") }()

	if c == nil {
		return errNoValue
	}

	var errs []error

	if c.Size == 0 {
		err = fmt.Errorf("got size %s: %w", c.Size, errMustBePositive)
		errs = append(errs, err)
	}

	if c.ClientSize == 0 {
		err = fmt.Errorf("got client_size %s: %w", c.ClientSize, errMustBePositive)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
