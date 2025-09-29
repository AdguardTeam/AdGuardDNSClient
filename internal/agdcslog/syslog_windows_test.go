//go:build windows

package agdcslog_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystemLogger_integration(t *testing.T) {
	requireIntegration(t)

	l := integrationSystemLogger(t)
	require.NotNil(t, l)
}
