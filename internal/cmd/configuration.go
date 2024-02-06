package cmd

import (
	"os"

	"github.com/AdguardTeam/golibs/errors"
	"gopkg.in/yaml.v3"
)

// configuration is the structure of YAML configuration for AdGuardDNSClient.
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

// parseConfiguration parses the YAML configuration file located at path.
func parseConfiguration(path string) (conf *configuration, err error) {
	// #nosec G304 -- Trust the path to the configuration file that is given
	// in the constant.
	f, err := os.Open(path)
	if err != nil {
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
