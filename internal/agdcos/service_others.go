//go:build !darwin

package agdcos

// validateExecPath is a no-op on non-Darwin platforms.
func validateExecPath(_ string) (err error) {
	return nil
}
