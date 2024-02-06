package cmd

// debugConfig is the configuration for debugging features.
type debugConfig struct {
	// Pprof configures profiling of the application.
	Pprof *pprofConfig `yaml:"pprof"`
}

// pprofConfig is the configuration for Go-provided runtime profiling tool.
type pprofConfig struct {
	// Port is used to serve debug HTTP API.
	Port uint16 `yaml:"port"`

	// Enabled specifies if the profiling enabled.
	Enabled bool `yaml:"enabled"`
}
