package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/AdguardTeam/golibs/osutil"
	osservice "github.com/kardianos/service"
)

// options specifies the command-line options.
//
// TODO(e.burkov):  Add an option to prevalidate configuration file.
type options struct {
	// serviceAction specifies the action to perform with the service.  See
	// [serviceAction] for the list of possible values.
	serviceAction serviceAction

	// verbose specifies whether to enable verbose output.
	verbose bool

	// help makes the application print the usage message and exit.
	help bool
}

// statusArgumentError is returned by AdGuardDNSClient when the program exits
// due to invalid command-line argument or its value.
const statusArgumentError = 2

// parseOptions parses the command-line options.
func parseOptions() (opts *options, err error) {
	const (
		optionService      = "s"
		descriptionService = "service action to perform, one of: " +
			string(serviceActionStart) + ", " +
			string(serviceActionStop) + ", " +
			string(serviceActionRestart) + ", " +
			string(serviceActionInstall) + ", " +
			string(serviceActionUninstall)

		optionVerbose      = "v"
		descriptionVerbose = "enable verbose logging"

		optionHelp      = "h"
		descriptionHelp = "print this help"
	)

	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	opts = &options{}

	flag.Var(&opts.serviceAction, optionService, descriptionService)
	flag.BoolVar(&opts.verbose, optionVerbose, false, descriptionVerbose)
	flag.BoolVar(&opts.help, optionHelp, false, descriptionHelp)

	return opts, flag.CommandLine.Parse(os.Args[1:])
}

// processOptions returns an [osservice.Service] to run.  It may appear nil when
// the program should exit immediately with exitCode.
func processOptions(opts *options, parseErr error) (svc osservice.Service, exitCode int) {
	if parseErr != nil {
		// Already reported by flag package.
		return nil, statusArgumentError
	}

	if opts.help {
		flag.CommandLine.SetOutput(os.Stdout)
		flag.CommandLine.Usage()

		return nil, osutil.ExitCodeSuccess
	}

	svc, confPath, err := newDefaultService(opts)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		return nil, osutil.ExitCodeFailure
	}

	if opts.serviceAction != "" {
		err = control(svc, confPath, opts.serviceAction)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)

			return nil, osutil.ExitCodeFailure
		}

		return nil, osutil.ExitCodeSuccess
	}

	return svc, 0
}
