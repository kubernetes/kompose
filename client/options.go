package client

// Opt is a configuration option to initialize a client
type Opt func(*Kompose) error

func WithSuppressWarnings() Opt {
	return func(k *Kompose) error {
		k.suppressWarnings = true
		return nil
	}
}

func WithVerboseOutput() Opt {
	return func(k *Kompose) error {
		k.verbose = true
		return nil
	}
}

func WithErrorOnWarning() Opt {
	return func(k *Kompose) error {
		k.errorOnWarning = true
		return nil
	}
}
