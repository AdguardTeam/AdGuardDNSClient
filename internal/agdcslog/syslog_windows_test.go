//go:build windows

package agdcslog_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/AdguardTeam/golibs/osutil/executil"
)

// cmdLogReader is the name of the system log reader.
const cmdLogReader = "wevtutil"

// findInLog searches the Windows Application event log for the message.
func findInLog(ctx context.Context, _, msg string) (ok bool, err error) {
	var stdOut, stdErr bytes.Buffer

	ms := testTimeout.Milliseconds()
	query := fmt.Sprintf(`*[System[TimeCreated[timediff(@SystemTime) <= %d]]]`, ms)

	err = executil.Run(ctx, executil.SystemCommandConstructor{}, &executil.CommandConfig{
		Path:   cmdLogReader,
		Args:   []string{"qe", "Application", "/rd:true", "/f:text", "/q:" + query},
		Stdout: &stdOut,
		Stderr: &stdErr,
	})
	if err != nil {
		return false, fmt.Errorf("log search failed: %w; stderr=%q", err, &stdErr)
	}

	return strings.Contains(stdOut.String(), msg), nil
}
