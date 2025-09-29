//go:build linux

package agdcslog_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AdguardTeam/golibs/osutil/executil"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemLogger_integration(t *testing.T) {
	requireIntegration(t)

	l := integrationSystemLogger(t)

	since := time.Now()
	sinceStr := since.Format(time.DateTime)

	msg := fmt.Sprintf("%d", since.UnixNano())

	ctx := testutil.ContextWithTimeout(t, testTimeout)
	l.DebugContext(ctx, msg)

	require.EventuallyWithT(t, func(ct *assert.CollectT) {
		findCtx, cancel := context.WithTimeout(ctx, testTimeout)
		defer cancel()

		ok, err := findInJournald(findCtx, sinceStr, msg)
		require.NoError(ct, err)

		assert.True(ct, ok)
	}, testTimeout, testTimeout/10)
}

func findInJournald(ctx context.Context, since, msg string) (ok bool, err error) {
	var stdOut, stdErr bytes.Buffer

	err = executil.Run(ctx, executil.SystemCommandConstructor{}, &executil.CommandConfig{
		Path:   "journalctl",
		Args:   []string{"-o", "cat", "-t", testServiceName, "--since", since, "--no-pager"},
		Stdout: &stdOut,
		Stderr: &stdErr,
	})
	if err != nil {
		return false, fmt.Errorf("journalctl failed: %w; stderr=%q", err, &stdErr)
	}

	return strings.Contains(stdOut.String(), msg), nil
}
