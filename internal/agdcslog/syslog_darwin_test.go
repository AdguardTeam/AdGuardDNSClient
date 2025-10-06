//go:build darwin

package agdcslog_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/AdguardTeam/golibs/osutil/executil"
)

// cmdLogReader is the name of the system log reader.
const cmdLogReader = "log"

// findInLog searches the macOS unified log for the message.
func findInLog(ctx context.Context, msg string) (ok bool, err error) {
	var stdOut, stdErr bytes.Buffer

	err = executil.Run(ctx, executil.SystemCommandConstructor{}, &executil.CommandConfig{
		Path:   cmdLogReader,
		Args:   []string{"show", "--style", "syslog", "--last", testTimeout.String(), "--info"},
		Stdout: &stdOut,
		Stderr: &stdErr,
	})
	if err != nil {
		return false, fmt.Errorf("log search failed: %w; stderr=%q", err, &stdErr)
	}

	return strings.Contains(stdOut.String(), msg), nil
}

func TestSystemLogger_integration(t *testing.T) {
	requireEventuallyFound(t, cmdLogReader, findInLog)
}
