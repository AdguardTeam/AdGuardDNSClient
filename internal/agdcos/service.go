// Package agdcos contains utilities for functions requiring system calls and
// other OS-specific APIs.
package agdcos

// ValidateExecPath returns an error if the path to the executable is not valid
// for the current platform.
func ValidateExecPath(execPath string) (err error) {
	return validateExecPath(execPath)
}
