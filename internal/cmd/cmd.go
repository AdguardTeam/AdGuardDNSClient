// Package cmd is the AdGuardDNSClient entry point.
package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/AdguardTeam/golibs/osutil"
	osservice "github.com/kardianos/service"
)

// Main is the entrypoint of AdGuardDNSClient.  Main may accept arguments, such
// as embedded assets and command-line arguments.
func Main() {
	// TODO(e.burkov):  Parse environments for earlier logging.

	opts, err := parseOptions()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		flag.CommandLine.SetOutput(os.Stderr)
		flag.CommandLine.Usage()

		os.Exit(statusArgumentError)
	}

	if opts.help {
		flag.CommandLine.SetOutput(os.Stdout)
		flag.CommandLine.Usage()

		os.Exit(osutil.ExitCodeSuccess)
	}

	conf, err := handleServiceConfig(opts.serviceAction)
	check(err)

	prog := &program{
		conf:  conf,
		done:  make(chan struct{}),
		errCh: make(chan error),
		log:   newLogger(opts, conf),
	}

	svc, err := osservice.New(prog, newServiceConfig())
	check(err)

	if opts.serviceAction != "" {
		err = control(svc, opts.serviceAction)
		check(err)

		os.Exit(osutil.ExitCodeSuccess)
	}

	err = svc.Run()
	check(err)
}

// check writes the error to stderr and exits the process with a failure code if
// the error is not nil.  It must only be called within [Main].
func check(err error) {
	if err == nil {
		return
	}

	// TODO(e.burkov):  Log errors properly.
	_, _ = fmt.Fprintln(os.Stderr, err)

	os.Exit(osutil.ExitCodeFailure)
}
