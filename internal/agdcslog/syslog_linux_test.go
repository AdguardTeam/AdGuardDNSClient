//go:build linux

package agdcslog_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/AdguardTeam/golibs/osutil/executil"
)

// cmdLogReader is the name of the system log reader.
const cmdLogReader = "journalctl"

// findInLog searches the systemd journal for the message.
func findInLog(ctx context.Context, msg string) (ok bool, err error) {
	var stdOut, stdErr bytes.Buffer

	since := "-" + testTimeout.String()

	err = executil.Run(ctx, executil.SystemCommandConstructor{}, &executil.CommandConfig{
		Path:   cmdLogReader,
		Args:   []string{"-o", "cat", "-t", testServiceName, "--since", since, "--no-pager"},
		Stdout: &stdOut,
		Stderr: &stdErr,
	})
	if err != nil {
		return false, fmt.Errorf("log search failed: %w; stderr=%q", err, &stdErr)
	}

	return strings.Contains(stdOut.String(), msg), nil
}
