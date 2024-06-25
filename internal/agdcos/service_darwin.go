//go:build darwin

package agdcos

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AdguardTeam/golibs/errors"
)

// errBadExecPath is returned when the executable is installed into an invalid
// location.
const errBadExecPath errors.Error = "bad executable path for service"

// validateExecPath returns an error if execPath is not a valid executable's
// location, i.e. is not within the /Applications directory.
//
// TODO(e.burkov):  Consider allowing the executable to be installed in other
// directories owned by root.
func validateExecPath(execPath string) (err error) {
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("evaluating executable path symlinks: %v", err)
	}

	execPath, err = filepath.Abs(execPath)
	if err != nil {
		return fmt.Errorf("getting absolute path of %q: %v", execPath, err)
	}

	if !strings.HasPrefix(execPath, "/Applications/") {
		return errBadExecPath
	}

	return nil
}
