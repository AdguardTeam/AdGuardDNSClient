package cmd

import (
	"flag"
	"os"
)

// options specifies the command-line options.
type options struct {
	// verbose specifies whether to enable verbose output.
	verbose bool

	// help makes the application print the usage message and exit.
	help bool
}

// Exit status constants.
const (
	statusSuccess       = 0
	statusError         = 1
	statusArgumentError = 2
)

// parseOptions parses the command-line options.
func parseOptions() (opts *options, err error) {
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	opts = &options{}

	flag.BoolVar(&opts.verbose, "v", false, "enable verbose logging")
	flag.BoolVar(&opts.help, "h", false, "print this help")

	return opts, flag.CommandLine.Parse(os.Args[1:])
}

// processOptions decides if AdGuardDNSClient should exit depending on the
// results of command-line option parsing.
func processOptions(opts *options, parseErr error) (exitCode int, needsExit bool) {
	if parseErr != nil {
		return statusArgumentError, true
	}

	if opts.help {
		flag.CommandLine.SetOutput(os.Stdout)
		flag.CommandLine.Usage()

		return statusSuccess, true
	}

	return 0, false
}
