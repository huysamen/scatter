package scatter

import (
	"context"
	"sync"
	"time"
)

// Error provides a wrapper struct around an error object so that we can include the original input with the error. This
// is due to not being able to guarantee the order in which tasks are executed in a goroutine pool.
type Error[I any] struct {
	Input I
	Error error
}

// RunFn specifies the signature for the function being passed into the pool which will get executed for each input.
type RunFn[I any, O any] func(I) (O, error)

// Run creates a specified number of goroutines and executes a batch on inputs concurrently as well as aggregates the
// results and/or errors.
func Run[I any, O any](numRoutines uint, inputs []I, fn RunFn[I, O]) (results []O, errors []Error[I]) {
	return run(context.Background(), numRoutines, inputs, fn)
}

// RunCtx creates specified number of goroutines and executes a batch on inputs concurrently as well as aggregates the
// results and/or errors. In addition, it takes a context which when the deadline is received, executions will stop and
// timeout errors returned for the remaining jobs.
func RunCtx[I any, O any](ctx context.Context, numRoutines uint, inputs []I, fn RunFn[I, O]) (results []O, errors []Error[I]) {
	return run(ctx, numRoutines, inputs, fn)
}

// RunWithDeadline is a helper function to run the inputs with a given deadline.
func RunWithDeadline[I any, O any](deadline time.Time, numRoutines uint, inputs []I, fn RunFn[I, O]) (results []O, errors []Error[I]) {
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	return run(ctx, numRoutines, inputs, fn)
}

// RunWithTimeout is a helper function to run the inputs with a given timeout.
func RunWithTimeout[I any, O any](timeout time.Duration, numRoutines uint, inputs []I, fn RunFn[I, O]) (results []O, errors []Error[I]) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return run(ctx, numRoutines, inputs, fn)
}

func run[I any, O any](ctx context.Context, numRoutines uint, inputs []I, fn RunFn[I, O]) (results []O, errors []Error[I]) {
	var wg sync.WaitGroup

	// create channels for the inputs, outputs as well as any errors that get generated
	ic := make(chan I, numRoutines+1)
	oc := make(chan O, len(inputs)+1)
	ec := make(chan Error[I], len(inputs)+1)

	// limit the wait group to the number provided by the caller
	wg.Add(int(numRoutines))

	// create the routines
	for i := uint(0); i < numRoutines; i++ {
		go func() {
			for {
				// grab the next input from the channel, stopping if the channel has been closed
				in, ok := <-ic
				if !ok {
					wg.Done()

					return
				}

				// check that the context has not been canceled or the deadline reached
				select {
				case <-ctx.Done():
					ec <- Error[I]{Input: in, Error: ctx.Err()}
				default:
					// execute the actual provided function
					out, err := fn(in)
					if err != nil {
						// wrap the error to include the input for identification purposes on the caller's side
						ec <- Error[I]{Input: in, Error: err}
					} else {
						oc <- out
					}
				}
			}
		}()
	}

	// pump all the inputs into the input channel so that they can be executed
	for _, in := range inputs {
		ic <- in
	}

	// close the channel to indicate all inputs have been taken
	close(ic)

	// wait for all jobs to finish
	wg.Wait()

	// close the output and error channels
	close(oc)
	close(ec)

	// aggregate all successful outputs
	for out := range oc {
		results = append(results, out)
	}

	// aggregate all errors
	for e := range ec {
		errors = append(errors, e)
	}

	return
}
