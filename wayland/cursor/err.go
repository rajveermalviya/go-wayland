package cursor

type MultiError struct {
	errs []error
}

func (e *MultiError) Err() error {
	if len(e.errs) == 0 {
		return nil
	}
	return e
}

func (e *MultiError) Add(err error) {
	if err == nil {
		return
	}
	e.errs = append(e.errs, err)
}

// Error returns `\n` separated string of all errors.
func (e *MultiError) Error() string {
	s := ""

	for _, err := range e.errs {
		s += (err.Error() + "\n")
	}

	return s
}
