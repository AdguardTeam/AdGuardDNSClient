// Package cmd is the AdGuardDNSClient entry point.
package cmd

import (
	"os"
)

// Main is the entrypoint of AdGuardDNSClient.  Main may accept arguments, such
// as embedded assets and command-line arguments.
func Main() {
	// TODO(e.burkov):  Parse environments for earlier logging.

	opts, err := parseOptions()
	svc, exitCode := processOptions(opts, err)
	if svc == nil {
		os.Exit(exitCode)
	}

	if err = svc.Run(); err != nil {
		// TODO(e.burkov):  Log errors properly.
		panic(err)
	}
}
