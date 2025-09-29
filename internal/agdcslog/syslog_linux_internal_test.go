//go:build linux

package agdcslog

import (
	"io"
	"log/syslog"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemLogger(t *testing.T) {
	t.Parallel()

	const (
		network = "tcp"

		testTag = "tag"

		debugMsg = "debug_msg"
		infoMsg  = "info_msg"
		warnMsg  = "warn_msg"
		errMsg   = "err_msg"
	)

	localhostAnyPort := netip.AddrPortFrom(netutil.IPv4Localhost(), 0)
	addr := net.TCPAddrFromAddrPort(localhostAnyPort)
	ln, err := net.ListenTCP(network, addr)
	require.NoError(t, err)
	testutil.CleanupAndRequireSuccess(t, ln.Close)

	w, err := syslog.Dial(network, ln.Addr().String(), syslog.LOG_LOCAL0, testTag)
	require.NoError(t, err)

	l := &systemLogger{writer: w}
	require.NoError(t, l.Debug(debugMsg))
	require.NoError(t, l.Info(infoMsg))
	require.NoError(t, l.Warning(warnMsg))
	require.NoError(t, l.Error(errMsg))

	// Close the syslog writer to signal EOF to the reader.
	require.NoError(t, l.Close())

	conn, err := ln.Accept()
	require.NoError(t, err)
	testutil.CleanupAndRequireSuccess(t, conn.Close)

	err = conn.SetReadDeadline(time.Now().Add(testTimeout))
	require.NoError(t, err)

	data, err := io.ReadAll(conn)
	require.NoError(t, err)

	out := string(data)
	assert.Contains(t, out, debugMsg)
	assert.Contains(t, out, infoMsg)
	assert.Contains(t, out, warnMsg)
	assert.Contains(t, out, errMsg)
}
