package cmd

// logConfig is the configuration for logging.
type logConfig struct {
	// File is the file to write logs to.  If empty, logs are written to stdout.
	File string `yaml:"file"`

	// Verbose specifies whether to log extra information.
	Verbose bool `yaml:"verbose"`
}
