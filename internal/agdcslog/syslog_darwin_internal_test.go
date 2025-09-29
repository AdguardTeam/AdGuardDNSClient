//go:build darwin

package agdcslog

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"

	"github.com/AdguardTeam/golibs/osutil/executil"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/AdguardTeam/golibs/testutil/fakeos/fakeexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lockedWriter protects concurrent writes to an underlying writer.
type lockedWriter struct {
	mu *sync.Mutex
	w  io.Writer
}

// type check
var _ io.Writer = (*lockedWriter)(nil)

// Write implements the [io.Writer] interface for *lockedWriter.
func (lw *lockedWriter) Write(b []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	return lw.w.Write(b)
}

func TestSystemLogger(t *testing.T) {
	t.Parallel()

	const (
		testTag = "tag"

		debugMsg = "debug_msg"
		infoMsg  = "info_msg"
		warnMsg  = "warn_msg"
		errMsg   = "err_msg"
	)

	stdoutBuf := bytes.Buffer{}
	lw := &lockedWriter{
		mu: &sync.Mutex{},
		w:  &stdoutBuf,
	}

	onNew := func(_ context.Context, conf *executil.CommandConfig) (c executil.Command, err error) {
		cmd := fakeexec.NewCommand()

		done := make(chan struct{})
		cmd.OnStart = func(_ context.Context) (err error) {
			go func(r io.Reader) {
				_, err = io.Copy(lw, r)
				require.NoError(testutil.PanicT{}, err)

				close(done)
			}(conf.Stdin)

			return nil
		}

		cmd.OnWait = func(_ context.Context) (err error) {
			<-done

			return nil
		}

		return cmd, nil
	}

	cmdCons := &fakeexec.CommandConstructor{
		OnNew: onNew,
	}

	ctx := testutil.ContextWithTimeout(t, testTimeout)
	l, err := newSystemLoggerWithCommandConstructor(ctx, cmdCons, testTag)
	require.NoError(t, err)

	require.NoError(t, l.Debug(debugMsg))
	require.NoError(t, l.Info(infoMsg))
	require.NoError(t, l.Warning(warnMsg))
	require.NoError(t, l.Error(errMsg))
	require.NoError(t, l.Close())

	out := stdoutBuf.String()
	assert.Contains(t, out, testTag+": "+debugMsg+"\n")
	assert.Contains(t, out, testTag+": "+infoMsg+"\n")
	assert.Contains(t, out, testTag+": "+warnMsg+"\n")
	assert.Contains(t, out, testTag+": "+errMsg+"\n")
}
