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

	// errMustBeUnique signals that a value must be unique.
	errMustBeUnique errors.Error = "must be unique"

	// errUnknownAction signals that an unknown service action was requested.
	errUnknownAction errors.Error = "unknown action"
)

// validator is a configuration object that is able to validate itself.
//
// TODO(e.burkov):  Think of a way to generalize slice validations.
//
// TODO(e.burkov):  Flatten the error messages.
type validator interface {
	// validate should return an error if the object considers itself invalid.
	// The error should contains the configuration's field.
	validate() (err error)
}
