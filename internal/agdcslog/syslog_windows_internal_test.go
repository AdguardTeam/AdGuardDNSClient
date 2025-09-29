//go:build windows

package agdcslog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeEventlogWriter is a mock implementation of the [eventlogWriter]
// interface.  It stores the provided messages and event IDs in its fields.
type fakeEventlogWriter struct {
	debugMsg string
	infoMsg  string
	warnMsg  string
	errMsg   string
	debugID  uint32
	infoID   uint32
	warnID   uint32
	errID    uint32
	closed   bool
}

// type check
var _ eventlogWriter = (*fakeEventlogWriter)(nil)

// Info implements the [eventlogWriter] interface for *fakeEventlogWriter.
func (f *fakeEventlogWriter) Info(id uint32, msg string) (err error) {
	if id == debugEventID {
		f.debugID, f.debugMsg = id, msg
	} else {
		f.infoID, f.infoMsg = id, msg
	}

	return nil
}

// Warning implements the [eventlogWriter] interface for *fakeEventlogWriter.
func (f *fakeEventlogWriter) Warning(id uint32, msg string) (err error) {
	f.warnID, f.warnMsg = id, msg

	return nil
}

// Error implements the [eventlogWriter] interface for *fakeEventlogWriter.
func (f *fakeEventlogWriter) Error(id uint32, msg string) (err error) {
	f.errID, f.errMsg = id, msg

	return nil
}

// Close implements the [eventlogWriter] interface for *fakeEventlogWriter.
func (f *fakeEventlogWriter) Close() error {
	f.closed = true

	return nil
}

func TestSystemLogger(t *testing.T) {
	t.Parallel()

	const (
		debugMsg = "debug_msg"
		infoMsg  = "info_msg"
		warnMsg  = "warn_msg"
		errMsg   = "err_msg"
	)

	f := &fakeEventlogWriter{}
	l := &systemLogger{writer: f}

	require.NoError(t, l.Debug(debugMsg))
	require.NoError(t, l.Info(infoMsg))
	require.NoError(t, l.Warning(warnMsg))
	require.NoError(t, l.Error(errMsg))
	require.NoError(t, l.Close())

	assert.Equal(t, debugMsg, f.debugMsg)
	assert.Equal(t, infoMsg, f.infoMsg)
	assert.Equal(t, warnMsg, f.warnMsg)
	assert.Equal(t, errMsg, f.errMsg)

	assert.Equal(t, uint32(debugEventID), f.debugID)
	assert.Equal(t, uint32(infoEventID), f.infoID)
	assert.Equal(t, uint32(warningEventID), f.warnID)
	assert.Equal(t, uint32(errorEventID), f.errID)

	assert.True(t, f.closed)
}
