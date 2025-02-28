package errgroup

type ErrGroup struct {
	Errors []error
}

func (g *ErrGroup) Error() string {
	var errMessage string

	for i, e := range g.Errors {
		if i == 0 {
			errMessage = e.Error()
		}

		errMessage += ": " + e.Error()
	}

	return errMessage
}

func (g *ErrGroup) ErrOrNil() error {
	if g == nil {
		return nil
	}

	if len(g.Errors) == 0 {
		return nil
	}

	return g
}

func (g *ErrGroup) Add(e error) {
	g.Errors = append(g.Errors, e)
}
