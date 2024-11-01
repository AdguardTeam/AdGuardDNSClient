package cmd

// validator is a configuration object that is able to validate itself.
//
// TODO(e.burkov):  Think of a way to generalize slice validations.
//
// TODO(e.burkov):  Flatten the error messages.
type validator interface {
	// validate should return an error if the object considers itself invalid.
	validate() (err error)
}
