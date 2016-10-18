package parallel

import (
	"sync"
)

// Run executes the provided functions in parallel and collects any errors they return.
func Run(fns ...func() error) []error {
	wg := sync.WaitGroup{}
	errCh := make(chan error, len(fns))
	wg.Add(len(fns))
	for i := range fns {
		go func(i int) {
			if err := fns[i](); err != nil {
				errCh <- err
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	close(errCh)
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	return errs
}
