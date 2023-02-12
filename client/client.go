package client

type Kompose struct {
	suppressWarnings bool
	verbose          bool
	errorOnWarning   bool
}

func NewClient(ops ...Opt) (*Kompose, error) {
	k := &Kompose{
		suppressWarnings: false,
		verbose:          false,
		errorOnWarning:   false,
	}
	for _, op := range ops {
		if err := op(k); err != nil {
			return nil, err
		}
	}
	return k, nil
}
