package cmd

import (
	"fmt"
	"os"

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

// parseConfig parses the YAML configuration file located at path.
func parseConfig(path string) (conf *configuration, err error) {
	// #nosec G304 -- Trust the path to the configuration file that is given
	// in the constant.
	f, err := os.Open(path)
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
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
