package cmd

import (
	"log/slog"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/log"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
)

// logConfig is the configuration for logging.
type logConfig struct {
	// TODO(e.burkov):  Add logging to file if needed.

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

	return nil
}

// newLogger creates a new logger from opts and conf.
//
// TODO(s.chzhen):  Add envs.
func newLogger(opts *options, conf *configuration) (l *slog.Logger) {
	// logFormat is the format of the log messages.
	//
	// TODO(e.burkov):  Use [log/slog] in [dnsproxy] and make it configurable.
	//
	// TODO(e.burkov):  Add unmarshalling to [slogutil.Format].
	const logFormat slogutil.Format = slogutil.FormatAdGuardLegacy

	isVerbose := opts.verbose
	if conf != nil {
		isVerbose = isVerbose || conf.Log.Verbose
	}

	// TODO(e.burkov):  Configure timestamp.
	l = slogutil.New(&slogutil.Config{
		Format:  logFormat,
		Verbose: isVerbose,
	})
	if isVerbose {
		log.SetLevel(log.DEBUG)
	}

	// TODO(e.burkov): Configure the service logger.

	return l
}
