package cmd

import (
	"github.com/AdguardTeam/golibs/errors"
)

// debugConfig is the configuration for debugging features.
type debugConfig struct {
	// Pprof configures profiling of the application.
	Pprof *pprofConfig `yaml:"pprof"`
}

// type check
var _ validator = (*debugConfig)(nil)

// validate implements the [validator] interface for *pprofConfig.
func (c *debugConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "debug section: %w") }()

	if c == nil {
		return errNoValue
	}

	return c.Pprof.validate()
}

// pprofConfig is the configuration for Go-provided runtime profiling tool.
type pprofConfig struct {
	// Port is used to serve debug HTTP API.
	Port uint16 `yaml:"port"`

	// Enabled specifies if the profiling enabled.
	Enabled bool `yaml:"enabled"`
}

// type check
var _ validator = (*pprofConfig)(nil)

// validate implements the [validator] interface for *pprofConfig.
func (c *pprofConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "pprof section: %w") }()

	if c == nil {
		return errNoValue
	}

	return nil
}
