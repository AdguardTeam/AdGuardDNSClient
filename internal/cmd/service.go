package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/AdguardTeam/golibs/errors"
	osservice "github.com/kardianos/service"
)

// newServiceConfig creates a configuration that the OS service manager uses to
// control the service.
func newServiceConfig() (conf *osservice.Config) {
	return &osservice.Config{
		Name:        "AdGuardDNSClient",
		DisplayName: "AdGuardDNS Client",
		Description: "A DNS client for AdGuardDNS",
	}
}

// serviceAction is a type for service actions that only allows a predefined set
// of values.
type serviceAction string

// Available service commands.
const (
	serviceActionNone      serviceAction = ""
	serviceActionInstall   serviceAction = "install"
	serviceActionRestart   serviceAction = "restart"
	serviceActionStart     serviceAction = "start"
	serviceActionStatus    serviceAction = "status"
	serviceActionStop      serviceAction = "stop"
	serviceActionUninstall serviceAction = "uninstall"
)

// type check
var _ flag.Value = (*serviceAction)(nil)

// Set implements the [flag.Value] interface for [*serviceAction].
func (a *serviceAction) Set(value string) (err error) {
	switch sa := serviceAction(value); sa {
	case
		serviceActionInstall,
		serviceActionRestart,
		serviceActionStart,
		serviceActionStatus,
		serviceActionStop,
		serviceActionUninstall:
		*a = sa

		return nil
	default:
		return errUnknownAction
	}
}

// String implements the [flag.Value] interface for [serviceAction].
func (a serviceAction) String() (s string) { return string(a) }

// control performs the specified service action.  It mirrors the service logic
// from [service.Control], but returns better errors.
func control(svc osservice.Service, action serviceAction) (err error) {
	defer func() { err = errors.Annotate(err, "performing %q: %w", action) }()

	switch action {
	case serviceActionInstall:
		return svc.Install()
	case serviceActionRestart:
		return svc.Restart()
	case serviceActionStart:
		return svc.Start()
	case serviceActionStatus:
		return controlStatus(svc)
	case serviceActionStop:
		return svc.Stop()
	case serviceActionUninstall:
		return svc.Uninstall()
	default:
		panic(errUnknownAction)
	}
}

// controlStatus prints the status of the system service corresponding to svc.
// It returns an error if the appropriate exit code should be used.
func controlStatus(svc osservice.Service) (err error) {
	status, err := svc.Status()
	if err != nil {
		return fmt.Errorf("retrieving status: %w", err)
	}

	var msg string
	switch status {
	case osservice.StatusRunning:
		msg = "running"
	case osservice.StatusStopped:
		msg = "stopped"
	default:
		// Don't expect [osservice.StatusUnknown] here, since it's only returned
		// on error.
		//
		// TODO(e.burkov):  Consider panicking here.
		return fmt.Errorf("unexpected status %d", status)
	}

	_, _ = fmt.Fprintln(os.Stdout, msg)

	return nil
}
