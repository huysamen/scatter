package scatter

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
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

	results, errs := Run[in, out](8, jobs, func(job in) (out, error) {
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

func TestRun_AllErrors(t *testing.T) {
	type in struct{}
	type out struct{}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{})
	}

	results, errs := Run[in, out](8, jobs, func(job in) (out, error) {
		return out{}, errors.New("error")
	})

	if len(errs) != 100 {
		t.Errorf("should produce only errors, received: %d", len(errs))
	}

	if len(results) != 0 {
		t.Errorf("should produce no results, received: %d", len(results))
	}
}

func TestRun_HalfErrors(t *testing.T) {
	type in struct{ i int }

	type out struct{}

	var jobs []in

	for i := 0; i < 100; i++ {
		jobs = append(jobs, in{i})
	}

	results, errs := Run[in, out](8, jobs, func(job in) (out, error) {
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

func TestRunCtx_NoDeadline(t *testing.T) {
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

	results, errs := RunCtx[in, out](ctx, 8, jobs, func(job in) (out, error) {
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

func TestRunCtx_WithDeadline(t *testing.T) {
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

	_, errs := RunCtx[in, out](ctx, 8, jobs, func(job in) (out, error) {
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

func TestRunCtx_WithTimeout(t *testing.T) {
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

	_, errs := RunCtx[in, out](ctx, 8, jobs, func(job in) (out, error) {
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

func TestRunCtx_WithCancel(t *testing.T) {
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

	_, errs := RunCtx[in, out](ctx, 8, jobs, func(job in) (out, error) {
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

func TestRun_WithDeadline(t *testing.T) {
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

	_, errs := RunWithDeadline[in, out](time.Now().Add(200*time.Millisecond), 8, jobs, func(job in) (out, error) {
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

func TestRun_WithTimeout(t *testing.T) {
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

	_, errs := RunWithTimeout[in, out](200*time.Millisecond, 8, jobs, func(job in) (out, error) {
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
