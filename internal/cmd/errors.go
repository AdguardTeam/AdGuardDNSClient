package cmd

import "github.com/AdguardTeam/golibs/errors"

const (
	// errNoValue signals that a required part of configuration is not present.
	errNoValue errors.Error = "no value"

	// errEmptyValue signals that a required part of configuration is empty.
	errEmptyValue errors.Error = "value is empty"

	// errMustBePositive signals that a numeric value must be greater than zero.
	errMustBePositive errors.Error = "must be positive"

	// errMustHaveNoMatch signals that an upstream group must have no match
	// criteria.
	errMustHaveNoMatch errors.Error = "must have no match criteria"
)

// check is a simple error-checking helper.  It must only be used within Main.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
