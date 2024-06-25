package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdcos"
	"github.com/AdguardTeam/golibs/errors"
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
	SchemaVersion schemaVersion `yaml:"schema_version"`
}

// defaultConfigPath is the path to the configuration file.
//
// TODO(e.burkov):  Make configurable via flags or environment.
const defaultConfigPath = "config.yaml"

// absolutePaths return the default path to the configuration file.  It assumes
// that the configuration file is located in the same directory as the
// executable.
func absolutePaths() (execPath, confPath string, err error) {
	execPath, err = os.Executable()
	if err != nil {
		return "", "", fmt.Errorf("getting executable path: %w", err)
	}

	absExecPath, err := filepath.Abs(execPath)
	if err != nil {
		return "", "", fmt.Errorf("getting absolute path of %q: %w", execPath, err)
	}

	return absExecPath, filepath.Join(filepath.Dir(absExecPath), defaultConfigPath), nil
}

// handleServiceConfig returns the service configuration based on the specified
// [serviceAction].
func handleServiceConfig(action serviceAction) (conf *configuration, err error) {
	execPath, confPath, err := absolutePaths()
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return nil, err
	}

	switch action {
	case serviceActionNone:
		conf, err = parseConfig(confPath)
		if err != nil {
			// Don't wrap the error since it's informative enough as is.
			return nil, err
		}

		err = conf.validate()
		if err != nil {
			// Don't wrap the error since it's informative enough as is.
			return nil, err
		}
	case serviceActionInstall:
		err = agdcos.ValidateExecPath(execPath)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "service executable must be located in the /Applications/ directory or its subdirectories")

			// Don't wrap the error since it's informative enough as is.
			return nil, err
		}

		err = writeDefaultConfig(confPath)
		if err != nil {
			// Don't wrap the error since it's informative enough as is.
			return nil, err
		}
	default:
		// No service actions require configuration.
		return nil, nil
	}

	return conf, nil
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
var _ validator = (*configuration)(nil)

// validate implements the [validator] interface for *configuration.
func (c *configuration) validate() (err error) {
	defer func() { err = errors.Annotate(err, "configuration: %w") }()

	if c == nil {
		return errNoValue
	}

	err = c.SchemaVersion.validate()
	if err != nil {
		// Don't validate the rest of the configuration of invalid schema
		// version.
		return err
	}

	var errs []error
	for _, v := range []validator{
		c.DNS,
		c.Log,
		c.Debug,
	} {
		errs = append(errs, v.validate())
	}

	return errors.Join(errs...)
}

// schemaVersion is the type for the configuration structure revision.
//
// TODO(e.burkov):  Move to configmigrate package.
type schemaVersion uint

// currentSchemaVersion is the current version of the configuration structure.
const currentSchemaVersion schemaVersion = 1

// type check
var _ validator = (schemaVersion)(currentSchemaVersion)

// validate implements the [validator] interface for schemaVersion.
func (v schemaVersion) validate() (err error) {
	defer func() { err = errors.Annotate(err, "schema_version: %w", v) }()

	switch {
	case v == 0:
		return errMustBePositive
	case v > currentSchemaVersion:
		return fmt.Errorf("got %d, most recent is %d", currentSchemaVersion, v)
	default:
		return nil
	}
}
