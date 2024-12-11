package scatter

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunIO(t *testing.T) {
	type in struct {
		i int
	}

	type out struct {
		i int
		o int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	results, errs := RunIO[in, out](8, jobs, func(job in) (out, error) {
		return out{i: job.i, o: job.i * 2}, nil
	})

	if len(errs) != 0 {
		t.Errorf("should produce no errors, received: %d", len(errs))
	}

	if len(results) != 100 {
		t.Errorf("should produce 100 results, received: %d", len(results))
	}

	for _, r := range results {
		if r.i*2 != r.o {
			t.Errorf("input values should be doubled: i=%d o=%d", r.i, r.o)
		}
	}
}

func TestRunI(t *testing.T) {
	type in struct {
		i int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	out := 0

	errs := RunI[in](8, jobs, func(job in) error {
		out = job.i + 1
		return nil
	})

	if len(errs) != 0 {
		t.Errorf("should produce no errors, received: %d", len(errs))
	}

	if out == 0 {
		t.Errorf("out should have been changed, received: %d", out)
	}
}

func TestRunIO_AllErrors(t *testing.T) {
	type in struct{}
	type out struct{}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{})
	}

	results, errs := RunIO[in, out](8, jobs, func(job in) (out, error) {
		return out{}, errors.New("error")
	})

	if len(errs) != 100 {
		t.Errorf("should produce only errors, received: %d", len(errs))
	}

	if len(results) != 0 {
		t.Errorf("should produce no results, received: %d", len(results))
	}
}

func TestRunI_AllErrors(t *testing.T) {
	type in struct{}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{})
	}

	errs := RunI[in](8, jobs, func(job in) error {
		return errors.New("error")
	})

	if len(errs) != 100 {
		t.Errorf("should produce only errors, received: %d", len(errs))
	}
}

func TestRunIO_HalfErrors(t *testing.T) {
	type in struct{ i int }

	type out struct{}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	results, errs := RunIO[in, out](8, jobs, func(job in) (out, error) {
		if job.i%2 == 0 {
			return out{}, nil
		}

		return out{}, errors.New("error")
	})

	if len(errs) != 50 {
		t.Errorf("should produce exactly 50 errors, received: %d", len(errs))
	}

	if len(results) != 50 {
		t.Errorf("should produce exactly 50 results, received: %d", len(results))
	}
}

func TestRunI_HalfErrors(t *testing.T) {
	type in struct{ i int }

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	errs := RunI[in](8, jobs, func(job in) error {
		if job.i%2 == 0 {
			return nil
		}

		return errors.New("error")
	})

	if len(errs) != 50 {
		t.Errorf("should produce exactly 50 errors, received: %d", len(errs))
	}
}

func TestRunIOCtx_NoDeadline(t *testing.T) {
	type in struct {
		i int
	}

	type out struct {
		i int
		o int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	ctx := context.Background()

	results, errs := RunIOCtx[in, out](ctx, 8, jobs, func(job in) (out, error) {
		return out{i: job.i, o: job.i * 2}, nil
	})

	if len(errs) != 0 {
		t.Errorf("should produce no errors, received: %d", len(errs))
	}

	if len(results) != 100 {
		t.Errorf("should produce 100 results, received: %d", len(results))
	}

	for _, r := range results {
		if r.i*2 != r.o {
			t.Errorf("input values should be doubled: i=%d o=%d", r.i, r.o)
		}
	}
}

func TestRunICtx_NoDeadline(t *testing.T) {
	type in struct {
		i int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	ctx := context.Background()

	errs := RunICtx[in](ctx, 8, jobs, func(job in) error {
		return nil
	})

	if len(errs) != 0 {
		t.Errorf("should produce no errors, received: %d", len(errs))
	}
}

func TestRunIOCtx_WithDeadline(t *testing.T) {
	type in struct {
		i int
	}

	type out struct {
		i int
		o int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(200*time.Millisecond))
	defer cancel()

	_, errs := RunIOCtx[in, out](ctx, 8, jobs, func(job in) (out, error) {
		time.Sleep(100 * time.Millisecond)

		return out{i: job.i, o: job.i * 2}, nil
	})

	if len(errs) == 0 {
		t.Errorf("should produce some errors, received: %d", len(errs))
	}

	for _, e := range errs {
		if !errors.Is(e.Error, context.DeadlineExceeded) {
			t.Errorf("should be context.DeadlineExceeded error, received: %s", e.Error.Error())
		}
	}
}

func TestRunICtx_WithDeadline(t *testing.T) {
	type in struct {
		i int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(200*time.Millisecond))
	defer cancel()

	errs := RunICtx[in](ctx, 8, jobs, func(job in) error {
		time.Sleep(100 * time.Millisecond)

		return nil
	})

	if len(errs) == 0 {
		t.Errorf("should produce some errors, received: %d", len(errs))
	}

	for _, e := range errs {
		if !errors.Is(e.Error, context.DeadlineExceeded) {
			t.Errorf("should be context.DeadlineExceeded error, received: %s", e.Error.Error())
		}
	}
}

func TestRunIOCtx_WithTimeout(t *testing.T) {
	type in struct {
		i int
	}

	type out struct {
		i int
		o int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, errs := RunIOCtx[in, out](ctx, 8, jobs, func(job in) (out, error) {
		time.Sleep(100 * time.Millisecond)

		return out{i: job.i, o: job.i * 2}, nil
	})

	if len(errs) == 0 {
		t.Errorf("should produce some errors, received: %d", len(errs))
	}

	for _, e := range errs {
		if !errors.Is(e.Error, context.DeadlineExceeded) {
			t.Errorf("should be context.DeadlineExceeded error, received: %s", e.Error.Error())
		}
	}
}

func TestRunICtx_WithTimeout(t *testing.T) {
	type in struct {
		i int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	errs := RunICtx[in](ctx, 8, jobs, func(job in) error {
		time.Sleep(100 * time.Millisecond)

		return nil
	})

	if len(errs) == 0 {
		t.Errorf("should produce some errors, received: %d", len(errs))
	}

	for _, e := range errs {
		if !errors.Is(e.Error, context.DeadlineExceeded) {
			t.Errorf("should be context.DeadlineExceeded error, received: %s", e.Error.Error())
		}
	}
}

func TestRunIOCtx_WithCancel(t *testing.T) {
	type in struct {
		i int
	}

	type out struct {
		i int
		o int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	_, errs := RunIOCtx[in, out](ctx, 8, jobs, func(job in) (out, error) {
		time.Sleep(100 * time.Millisecond)

		return out{i: job.i, o: job.i * 2}, nil
	})

	if len(errs) == 0 {
		t.Errorf("should produce some errors, received: %d", len(errs))
	}

	for _, e := range errs {
		if !errors.Is(e.Error, context.Canceled) {
			t.Errorf("should be context.Canceled error, received: %s", e.Error.Error())
		}
	}
}

func TestRunICtx_WithCancel(t *testing.T) {
	type in struct {
		i int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	errs := RunICtx[in](ctx, 8, jobs, func(job in) error {
		time.Sleep(100 * time.Millisecond)

		return nil
	})

	if len(errs) == 0 {
		t.Errorf("should produce some errors, received: %d", len(errs))
	}

	for _, e := range errs {
		if !errors.Is(e.Error, context.Canceled) {
			t.Errorf("should be context.Canceled error, received: %s", e.Error.Error())
		}
	}
}

func TestRunIO_WithDeadline(t *testing.T) {
	type in struct {
		i int
	}

	type out struct {
		i int
		o int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	_, errs := RunIOWithDeadline[in, out](time.Now().Add(200*time.Millisecond), 8, jobs, func(job in) (out, error) {
		time.Sleep(100 * time.Millisecond)

		return out{i: job.i, o: job.i * 2}, nil
	})

	if len(errs) == 0 {
		t.Errorf("should produce some errors, received: %d", len(errs))
	}

	for _, e := range errs {
		if !errors.Is(e.Error, context.DeadlineExceeded) {
			t.Errorf("should be context.DeadlineExceeded error, received: %s", e.Error.Error())
		}
	}
}

func TestRunI_WithDeadline(t *testing.T) {
	type in struct {
		i int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	errs := RunIWithDeadline[in](time.Now().Add(200*time.Millisecond), 8, jobs, func(job in) error {
		time.Sleep(100 * time.Millisecond)

		return nil
	})

	if len(errs) == 0 {
		t.Errorf("should produce some errors, received: %d", len(errs))
	}

	for _, e := range errs {
		if !errors.Is(e.Error, context.DeadlineExceeded) {
			t.Errorf("should be context.DeadlineExceeded error, received: %s", e.Error.Error())
		}
	}
}

func TestRunIO_WithTimeout(t *testing.T) {
	type in struct {
		i int
	}

	type out struct {
		i int
		o int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	_, errs := RunIOWithTimeout[in, out](200*time.Millisecond, 8, jobs, func(job in) (out, error) {
		time.Sleep(100 * time.Millisecond)

		return out{i: job.i, o: job.i * 2}, nil
	})

	if len(errs) == 0 {
		t.Errorf("should produce some errors, received: %d", len(errs))
	}

	for _, e := range errs {
		if !errors.Is(e.Error, context.DeadlineExceeded) {
			t.Errorf("should be context.DeadlineExceeded error, received: %s", e.Error.Error())
		}
	}
}

func TestRunI_WithTimeout(t *testing.T) {
	type in struct {
		i int
	}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	errs := RunIWithTimeout[in](200*time.Millisecond, 8, jobs, func(job in) error {
		time.Sleep(100 * time.Millisecond)

		return nil
	})

	if len(errs) == 0 {
		t.Errorf("should produce some errors, received: %d", len(errs))
	}

	for _, e := range errs {
		if !errors.Is(e.Error, context.DeadlineExceeded) {
			t.Errorf("should be context.DeadlineExceeded error, received: %s", e.Error.Error())
		}
	}
}

func TestRunI_WithPanic(t *testing.T) {
	type in struct {
		i int
	}

	var jobs []in

	for i := 0; i < 10; i++ {
		jobs = append(jobs, in{i})
	}

	errs := RunI[in](2, jobs, func(job in) error {
		panic("force panic")

		return nil
	})

	if len(errs) == 0 {
		t.Errorf("panic should produce errors, received: %d", len(errs))
	}

	if len(errs) != 10 {
		t.Errorf("panic should produce 10 errors, received: %d", len(errs))
	}
}

func TestRunIO_WithPanic(t *testing.T) {
	type in struct {
		i int
	}

	type out struct {
		i int
		o int
	}

	var jobs []in

	for i := 0; i < 10; i++ {
		jobs = append(jobs, in{i})
	}

	_, errs := RunIO[in, out](2, jobs, func(job in) (out, error) {
		panic("force panic")

		return out{}, nil
	})

	if len(errs) == 0 {
		t.Errorf("panic should produce errors, received: %d", len(errs))
	}

	if len(errs) != 10 {
		t.Errorf("panic should produce 10 errors, received: %d", len(errs))
	}
}
