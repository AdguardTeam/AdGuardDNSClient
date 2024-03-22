package cmd

import "github.com/AdguardTeam/golibs/errors"

// logConfig is the configuration for logging.
type logConfig struct {
	// File is the file to write logs to.  If empty, logs are written to stdout.
	File string `yaml:"file"`

	// Verbose specifies whether to log extra information.
	Verbose bool `yaml:"verbose"`
}

// type check
var _ validator = (*logConfig)(nil)

// validate implements the [validator] interface for *logConfig.
func (c *logConfig) validate() (err error) {
	defer func() { err = errors.Annotate(err, "log: %w") }()

	if c == nil {
		return errNoValue
	}

	// TODO(e.burkov):  Check the file path.

	return nil
}
