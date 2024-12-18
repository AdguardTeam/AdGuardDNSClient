package cmd

import (
	"fmt"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/validate"
)

// debugConfig is the configuration for debugging features.
type debugConfig struct {
	// Pprof configures profiling of the application.
	Pprof *pprofConfig `yaml:"pprof"`
}

// type check
var _ validate.Interface = (*debugConfig)(nil)

// Validate implements the [validate.Interface] interface for *debugConfig.
func (c *debugConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	err = c.Pprof.Validate()
	if err != nil {
		return fmt.Errorf("pprof: %w", err)
	}

	return nil
}

// pprofConfig is the configuration for Go-provided runtime profiling tool.
type pprofConfig struct {
	// Port is used to serve debug HTTP API.
	Port uint16 `yaml:"port"`

	// Enabled specifies if the profiling enabled.
	Enabled bool `yaml:"enabled"`
}

// type check
var _ validate.Interface = (*pprofConfig)(nil)

// Validate implements the [validate.Interface] interface for *pprofConfig.
func (c *pprofConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	return nil
}
