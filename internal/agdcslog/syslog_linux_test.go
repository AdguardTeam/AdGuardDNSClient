//go:build linux

package agdcslog_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/AdguardTeam/golibs/osutil/executil"
)

const (
	// systemdLogReader is the systemd log reader.
	systemdLogReader = "journalctl"

	// syslogLogReader is the BusyBox syslog log reader.
	syslogLogReader = "logread"
)

// findWithJournalctl searches the systemd journal for the message.
func findWithJournalctl(ctx context.Context, msg string) (ok bool, err error) {
	var stdOut, stdErr bytes.Buffer

	since := "-" + testTimeout.String()

	err = executil.Run(ctx, executil.SystemCommandConstructor{}, &executil.CommandConfig{
		Path:   systemdLogReader,
		Args:   []string{"-o", "cat", "-t", testServiceName, "--since", since, "--no-pager"},
		Stdout: &stdOut,
		Stderr: &stdErr,
	})
	if err != nil {
		return false, fmt.Errorf("log search failed: %w; stderr=%q", err, &stdErr)
	}

	return strings.Contains(stdOut.String(), msg), nil
}

// findWithLogread searches the syslog ring buffer for the message.
func findWithLogread(ctx context.Context, msg string) (ok bool, err error) {
	var stdOut, stdErr bytes.Buffer

	err = executil.Run(ctx, executil.SystemCommandConstructor{}, &executil.CommandConfig{
		Path:   syslogLogReader,
		Args:   nil,
		Stdout: &stdOut,
		Stderr: &stdErr,
	})
	if err != nil {
		return false, fmt.Errorf("log search failed: %w; stderr=%q", err, &stdErr)
	}

	return strings.Contains(stdOut.String(), msg), nil
}

func TestSystemLogger_integration_systemd(t *testing.T) {
	requireEventuallyFound(t, systemdLogReader, findWithJournalctl)
}

func TestSystemLogger_integration_syslogd(t *testing.T) {
	requireEventuallyFound(t, syslogLogReader, findWithLogread)
}
