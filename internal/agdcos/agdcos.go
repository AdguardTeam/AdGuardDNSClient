// Package agdcos contains utilities for functions requiring system calls and
// other OS-specific APIs.
package agdcos

import (
	"io/fs"
)

// Default file, binary, and directory permissions.
const (
	DefaultPermDir  fs.FileMode = 0o700
	DefaultPermExe  fs.FileMode = DefaultPermFile | 0o100
	DefaultPermFile fs.FileMode = 0o600
)
