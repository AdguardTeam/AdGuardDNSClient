package cmd

import (
	"cmp"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdcslog"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
)

// Constants that define the log output.
const (
	outputSyslog = "syslog"
	outputStdout = "stdout"
	outputStderr = "stderr"
)

// logConfig is the configuration for logging.
type logConfig struct {
	// Output specifies output for logs.  Value must be empty, an absolute path
	// to a file, or one of the special values:
	//
	//	- [outputSyslog]
	//	- [outputStdout]
	//	- [outputStderr]
	Output string `yaml:"output"`

	// Format specifies format for logs.  Value must be empty or a valid
	// [slogutil.Format].  Note that system log entries are in text format.
	Format slogutil.Format `yaml:"format"`

	// Timestamp specifies whether to add timestamp to the log entries.
	Timestamp bool `yaml:"timestamp"`

	// Verbose specifies whether to log extra information.
	Verbose bool `yaml:"verbose"`
}

// type check
var _ validator = (*logConfig)(nil)

// validate implements the [validator] interface for *logConfig.
func (c *logConfig) validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	var errs []error
	switch c.Output {
	case outputSyslog, outputStdout, outputStderr:
		// Go on.
	default:
		if !filepath.IsAbs(c.Output) {
			errs = append(errs, fmt.Errorf("unsupported log output: %q", c.Output))
		}
	}

	// TODO(e.burkov):  Add unmarshalling to [slogutil.Format].
	_, err = slogutil.NewFormat(string(c.Format))
	errs = append(errs, err)

	return errors.Join(errs...)
}

// newEnvLogger returns a new default logger using the information from the
// environment, cmdline arguments, and defaults.
func newEnvLogger(
	opts *options,
	envs *logEnvs,
) (l *slog.Logger, logFile *os.File, err error) {
	output := cmp.Or(envs.output, outputSyslog)
	format := cmp.Or(envs.format, slogutil.FormatDefault)
	isVerbose := opts.verbose || (envs.verboseSet && envs.verbose)

	return newLogger(output, format, envs.timestampSet && envs.timestamp, isVerbose)
}

// newLogger creates a new logger based on the parameters.  l is never nil: if
// the file or the system log cannot be opened, l writes to [os.Stderr].
func newLogger(
	outputStr string,
	f slogutil.Format,
	addTimestamp bool,
	isVerbose bool,
) (l *slog.Logger, logFile *os.File, err error) {
	var output *os.File
	if outputStr == outputSyslog {
		l, err = newSystemLogger(isVerbose)
		if err == nil {
			return l, nil, nil
		}

		err = fmt.Errorf("opening syslog: %w", err)
		output = os.Stderr
	} else {
		var needsClose bool
		output, needsClose, err = outputFromStr(outputStr)
		if err != nil {
			err = fmt.Errorf("opening log file: %w", err)
		} else if needsClose {
			logFile = output
		}
	}

	return slogutil.New(&slogutil.Config{
		Output:       output,
		Format:       f,
		AddTimestamp: addTimestamp,
		Verbose:      isVerbose,
	}), logFile, err
}

// outputFromStr converts a string, which must be one of [outputStderr],
// [outputStdout], and an absolute file path, into an output.  If the output
// requires closing, needsClose is true.  output is never nil: if the file
// cannot be opened, output is [os.Stderr].
func outputFromStr(s string) (output *os.File, needsClose bool, err error) {
	switch s {
	case outputStderr:
		output = os.Stderr
	case outputStdout:
		output = os.Stdout
	default:
		needsClose = true
		// #nosec G304 -- Trust the user provided path.
		output, err = os.OpenFile(s, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			output = os.Stderr
			needsClose = false
			err = fmt.Errorf("opening config output: %w", err)
		}
	}

	return output, needsClose, err
}

// newSystemLogger returns a new logger that writes to system log with the
// given verbosity.
func newSystemLogger(isVerbose bool) (l *slog.Logger, err error) {
	sl, err := agdcslog.NewSystemLogger(serviceName)
	if err != nil {
		return nil, err
	}

	lvl := slog.LevelInfo
	if isVerbose {
		lvl = slog.LevelDebug
	}

	h := agdcslog.NewSyslogHandler(sl, &slog.HandlerOptions{
		Level: lvl,
	})

	return slog.New(h), nil
}

// newConfigLogger returns a logger based on the configuration file and
// including overrides from the environment and options.
//
// TODO(a.garipov):  Refactor.
func newConfigLogger(
	envLogger *slog.Logger,
	envLogFile *os.File,
	opts *options,
	envs *logEnvs,
	conf *configuration,
) (l *slog.Logger, logFile *os.File, errs []error) {
	if conf == nil {
		return envLogger, envLogFile, nil
	}

	c := conf.Log

	outputStr, format, addTimestamp, isVerbose := overridenLogConf(opts, envs, c)

	// Select an action based on the previous output.
	var err error
	if envs.output == "" {
		// envLogger is likely a syslog one, because the output was unset in the
		// environment.  Either close it or use it, depending on the verbosity
		// parameter.
		var usePrev bool
		usePrev, err = closeEnv(envLogger, opts, envs, c)
		if usePrev {
			return envLogger, nil, nil
		}

		if err != nil {
			errs = append(errs, fmt.Errorf("closing env syslog: %w", err))
		}
	} else if envLogFile != nil {
		return slogutil.New(&slogutil.Config{
			Output:       envLogFile,
			Format:       format,
			AddTimestamp: addTimestamp,
			Verbose:      isVerbose,
		}), envLogFile, nil
	}

	l, logFile, err = newLogger(outputStr, format, addTimestamp, isVerbose)
	if err != nil {
		errs = append(errs, fmt.Errorf("creating conf log: %w", err))
	}

	return l, logFile, errs
}

// overridenLogConf returns logging parameters with overrides from cmdline
// options and environment.
func overridenLogConf(
	opts *options,
	envs *logEnvs,
	c *logConfig,
) (outputStr string, format slogutil.Format, addTimestamp, isVerbose bool) {
	outputStr = cmp.Or(envs.output, c.Output)
	format = cmp.Or(envs.format, c.Format)

	addTimestamp = c.Timestamp
	if envs.timestampSet {
		addTimestamp = envs.timestamp
	}

	isVerbose = c.Verbose
	if envs.verboseSet {
		isVerbose = envs.verboseSet && envs.verbose
	}

	isVerbose = isVerbose || opts.verbose

	return outputStr, format, addTimestamp, isVerbose
}

// closeEnv closes the previous logger, if necessary.  If usePrev is true, the
// previous logger should be used.
func closeEnv(
	envLogger *slog.Logger,
	opts *options,
	envs *logEnvs,
	c *logConfig,
) (usePrev bool, closeErr error) {
	// Don't reopen the syslog if it's the same one.  Otherwise, close and
	// create a new logger.
	//
	// TODO(a.garipov):  Update when syslog handler supports adding
	// timestamp.

	// Both -v in the cmdline options and a set VERBOSE env mean that the
	// verbosity is overridden and thus doesn't change.  If neither of these is
	// true, compare the verbosity in the configuration file to the default
	// value, false.
	verboseIsSame := !c.Verbose
	if opts.verbose || envs.verboseSet {
		verboseIsSame = true
	}

	if c.Output == outputSyslog && verboseIsSame {
		return true, nil
	}

	if closer, ok := envLogger.Handler().(io.Closer); ok {
		closeErr = closer.Close()
	}

	return false, closeErr
}
