package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AdguardTeam/golibs/logutil/slogutil"
)

// Constants that define the log environment variables
const (
	envLogOutput    = "LOG_OUTPUT"
	envLogFormat    = "LOG_FORMAT"
	envLogTimestamp = "LOG_TIMESTAMP"
	envLogVerbose   = "VERBOSE"
)

// logEnvs represents the configuration for logging that is kept in the
// environment.
type logEnvs struct {
	// output specifies output for logs.  Empty string indicates that
	// environment variable is not set.  Value must be empty, an absolute path
	// to a file, or one of the special values:
	//
	//	- [outputSyslog]
	//	- [outputStdout]
	//	- [outputStderr]
	output string

	// format specifies format for logs.  Empty string indicates that
	// environment variable is not set.  Value must be in [slogutil.Format].
	// Note that system log entries are in text format.
	format slogutil.Format

	// timestamp specifies whether to add timestamp to the log entries.
	timestamp bool

	// timestampSet indicates whether timestamp is set.
	//
	// TODO(s.chzhen):  Add NullBool to [golibs].
	timestampSet bool

	// verbose specifies whether to log extra information.
	verbose bool

	// verboseSet indicates whether verbose is set.
	//
	// TODO(s.chzhen):  Add NullBool to [golibs].
	verboseSet bool
}

// parseLogEnvs returns the log configuration read from the environment.  errs
// are the errors encountered while parsing.  envs is never nil.
//
// TODO(a.garipov):  Refactor.
func parseLogEnvs() (envs *logEnvs, errs []error) {
	envs = &logEnvs{}

	output := os.Getenv(envLogOutput)
	switch output {
	case outputStdout, outputStderr, outputSyslog, "":
		envs.output = output
	default:
		if !filepath.IsAbs(output) {
			errs = append(errs, fmt.Errorf("unsupported log output: %q", output))
		} else {
			envs.output = output
		}
	}

	format := os.Getenv(envLogFormat)
	if format != "" {
		f, err := slogutil.NewFormat(format)
		if err != nil {
			errs = append(errs, fmt.Errorf("parsing %q: %w", envLogFormat, err))
		} else {
			envs.format = f
		}
	}

	var err error
	envs.timestamp, envs.timestampSet, err = envBool(envLogTimestamp)
	if err != nil {
		errs = append(errs, fmt.Errorf("parsing %s: %w", envLogTimestamp, err))
	}

	envs.verbose, envs.verboseSet, err = envBool(envLogVerbose)
	if err != nil {
		errs = append(errs, fmt.Errorf("parsing %s: %w", envLogVerbose, err))
	}

	return envs, errs
}

// envBool returns a boolean value read from the environment or a parsing error.
// ok is false and err is nil if env is not present.
func envBool(env string) (value, ok bool, err error) {
	str := os.Getenv(env)
	if str == "" {
		return false, false, nil
	}

	val, err := strconv.ParseBool(str)
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return false, false, err
	}

	return val, true, nil
}
