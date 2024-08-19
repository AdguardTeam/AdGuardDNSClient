package cmd

import "github.com/AdguardTeam/golibs/errors"

const (
	// errMustHaveNoMatch signals that an upstream group must have no match
	// criteria.
	errMustHaveNoMatch errors.Error = "must have no match criteria"

	// errMustBeUnique signals that a value must be unique.
	errMustBeUnique errors.Error = "must be unique"
)

// validator is a configuration object that is able to validate itself.
//
// TODO(e.burkov):  Think of a way to generalize slice validations.
//
// TODO(e.burkov):  Flatten the error messages.
type validator interface {
	// validate should return an error if the object considers itself invalid.
	validate() (err error)
}
