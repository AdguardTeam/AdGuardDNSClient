package cmd

import (
	"fmt"
	"os"

	"github.com/AdguardTeam/golibs/errors"
	"gopkg.in/yaml.v3"
)

// currentSchemaVersion is the current version of the configuration structure.
const currentSchemaVersion = 1

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
	SchemaVersion int `yaml:"schema_version"`
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

// validator is a configuration object that is able to validate itself.
type validator interface {
	// validate should return an error if the object considers itself invalid.
	validate() (err error)
}

// type check
var _ validator = (*configuration)(nil)

// validate implements the [validator] interface for *configuration.
func (c *configuration) validate() (err error) {
	defer func() { err = errors.Annotate(err, "validating configuration: %w") }()

	switch {
	case c == nil:
		return errNoValue
	case c.SchemaVersion > currentSchemaVersion:
		return fmt.Errorf(
			"got schema version %d, most recent is %d",
			c.SchemaVersion,
			currentSchemaVersion,
		)
	case c.SchemaVersion <= 0:
		return fmt.Errorf("got schema version %d: %w", c.SchemaVersion, errMustBePositive)
	}

	validators := []validator{
		c.DNS,
		c.Log,
		c.Debug,
	}

	var errs []error
	for _, v := range validators {
		errs = append(errs, v.validate())
	}

	return errors.Join(errs...)
}
