//go:build windows

package agdcslog_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/AdguardTeam/golibs/osutil/executil"
)

// cmdLogReader is the name of the system log reader.
const cmdLogReader = "wevtutil"

// findInLog searches the Windows Application event log for the message.
func findInLog(ctx context.Context, msg string) (ok bool, err error) {
	var stdOut, stdErr bytes.Buffer

	ms := testTimeout.Milliseconds()
	query := fmt.Sprintf(`*[System[TimeCreated[timediff(@SystemTime) <= %d]]]`, ms)

	err = executil.Run(ctx, executil.SystemCommandConstructor{}, &executil.CommandConfig{
		Path:   cmdLogReader,
		Args:   []string{"qe", "Application", "/rd:true", "/f:text", "/uni", "/q:" + query},
		Stdout: &stdOut,
		Stderr: &stdErr,
	})
	if err != nil {
		return false, fmt.Errorf("log search failed: %w; stderr=%q", err, &stdErr)
	}

	// Strip NUL bytes from UTF-16LE output.
	out := stdOut.Bytes()
	out = bytes.ReplaceAll(out, []byte{0x00}, nil)

	return bytes.Contains(out, []byte(msg)), nil
}

func TestSystemLogger_integration(t *testing.T) {
	requireEventuallyFound(t, cmdLogReader, findInLog)
}
