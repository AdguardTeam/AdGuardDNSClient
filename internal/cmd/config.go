package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdcos"
	"github.com/AdguardTeam/AdGuardDNSClient/internal/configmigrate"
	"github.com/AdguardTeam/golibs/container"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/timeutil"
	"github.com/AdguardTeam/golibs/validate"
	"gopkg.in/yaml.v3"
)

// configuration is the structure of YAML configuration for AdGuardDNSClient.
//
// TODO(e.burkov):  Test it out.
type configuration struct {
	// DNS configures processing of DNS requests.
	DNS *dnsConfig `yaml:"dns"`

	// Debug configures debugging features.
	Debug *debugConfig `yaml:"debug"`

	// Log configures logging.
	Log *logConfig `yaml:"log"`

	// SchemaVersion is the current version of this structure.  This is bumped
	// each time the configuration changes breaking backwards compatibility.
	SchemaVersion configmigrate.SchemaVersion `yaml:"schema_version"`
}

// defaultConfigName is the path to the configuration file.
//
// TODO(e.burkov):  Make configurable via flags or environment.
const defaultConfigName = "config.yaml"

// absolutePaths return the default path to the configuration file.  It assumes
// that the configuration file is located in the same directory as the
// executable.
func absolutePaths() (execPath, workDir string, err error) {
	execPath, err = os.Executable()
	if err != nil {
		return "", "", fmt.Errorf("getting executable path: %w", err)
	}

	absExecPath, err := filepath.Abs(execPath)
	if err != nil {
		return "", "", fmt.Errorf("getting absolute path of %q: %w", execPath, err)
	}

	workDir = filepath.Dir(absExecPath)

	return absExecPath, workDir, nil
}

// handleServiceConfig returns the service configuration based on the specified
// [serviceAction].
func handleServiceConfig(
	ctx context.Context,
	l *slog.Logger,
	action serviceAction,
) (conf *configuration, err error) {
	execPath, workDir, err := absolutePaths()
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return nil, err
	}

	switch action {
	case serviceActionNone:
		return handleConfig(ctx, l, workDir)
	case serviceActionInstall:
		return nil, handleInstall(execPath, workDir)
	default:
		// No service actions require configuration.
		return nil, nil
	}
}

// handleConfig parses the configuration file located in the [workDir]
// directory.
func handleConfig(
	ctx context.Context,
	l *slog.Logger,
	workDir string,
) (conf *configuration, err error) {
	migrator := configmigrate.New(&configmigrate.Config{
		Clock:          timeutil.SystemClock{},
		Logger:         l.With(slogutil.KeyPrefix, "configmigrate"),
		WorkingDir:     workDir,
		ConfigFileName: defaultConfigName,
	})
	err = migrator.Run(ctx, configmigrate.VersionLatest)
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return nil, err
	}

	conf, err = parseConfig(filepath.Join(workDir, defaultConfigName))
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return nil, err
	}

	err = conf.Validate()
	if err != nil {
		return nil, fmt.Errorf("configuration: %w", err)
	}

	return conf, nil
}

// handleInstall creates and writes the default configuration file to the
// [workDir] directory.
func handleInstall(execPath, workDir string) (err error) {
	err = agdcos.ValidateExecPath(execPath)
	if err != nil {
		// locWarnMsg is a warning message that is printed to stderr when the
		// service executable is not located correctly.
		//
		// TODO(e.burkov):  Move the OS-specific message to agdcos package,
		// perhaps, add a structured error with text.
		const locWarnMsg = "service executable must be located " +
			"in the /Applications/ directory or its subdirectories"

		_, _ = fmt.Fprintln(os.Stderr, locWarnMsg)

		// Don't wrap the error since it's informative enough as is.
		return err
	}

	err = writeDefaultConfig(filepath.Join(workDir, defaultConfigName))
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return err
	}

	return nil
}

// parseConfig parses the YAML configuration file located at path.
func parseConfig(path string) (conf *configuration, err error) {
	defer func() { err = errors.Annotate(err, "parsing configuration: %w") }()

	// #nosec G304 -- Trust the path to the configuration file that is currently
	// expected to be in the same directory as the binary.
	f, err := os.Open(path)
	if err != nil {
		// Don't wrap the error since there is already an annotation deferred.
		return nil, err
	}
	defer func() { err = errors.WithDeferred(err, f.Close()) }()

	conf = &configuration{}
	err = yaml.NewDecoder(f).Decode(conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// type check
var _ validate.Interface = (*configuration)(nil)

// Validate implements the [validate.Interface] interface for *configuration.
func (c *configuration) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	validators := container.KeyValues[string, validate.Interface]{{
		Key:   "dns",
		Value: c.DNS,
	}, {
		Key:   "log",
		Value: c.Log,
	}, {
		Key:   "debug",
		Value: c.Debug,
	}}

	var errs []error
	for _, v := range validators {
		errs = validate.Append(errs, v.Key, v.Value)
	}

	return errors.Join(errs...)
}
