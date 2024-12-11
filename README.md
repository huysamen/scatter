# Scatter
Simple library for running multiple jobs concurrently using goroutines.

### Installation

To install the Scatter library, use the following command:

```sh
go get github.com/huysamen/scatter
```

### Usage

The Scatter library provides a simple API for running multiple jobs concurrently using goroutines. It provides
support for running functions that only take input (and produces no output), as well as functions that take
input and produce output.

The library also provides support for dealing with contexts with cancellation, deadlines and timeouts.

Lastly, as this library makes use of goroutines, it also handles panic recovery and will return any panics as
errors.


### Examples

For complete examples, see the [test](scatter_test.go) file.

Here is an example of how to use the Scatter library to run multiple jobs concurrently, whilst
dealing with context cancellation.

```go
package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/huysamen/scatter"
)

type in struct {
	i int
}

type out struct {
	i int
	o int
}

func main() {
	var jobs []in
	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	results, errs := scatter.RunIOCtx[in, out](ctx, 8, jobs, func(job in) (out, error) {
		time.Sleep(100 * time.Millisecond)
		
		return out{i: job.i, o: job.i * 2}, nil
	})

	fmt.Println("Results:", results)
	fmt.Println("Errors:", errs)
}
```
