package processor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewPoolThresholds(t *testing.T) {
	p := NewPool(2, 8, 20)
	require.Equal(t, 2, p.min)
	require.Equal(t, 8, p.max)
	require.Equal(t, 10, p.upAt)  // 20 * 1 / 2
	require.Equal(t, 2, p.downAt) // 20 / 10
	require.NotNil(t, p.inputCh)
}

func TestEnqueueSuccess(t *testing.T) {
	p := NewPool(1, 2, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calls := int32(0)
	job := func(context.Context) { atomic.AddInt32(&calls, 1) }

	go func() { _ = p.Run(ctx) }()
	time.Sleep(30 * time.Millisecond)

	require.NoError(t, p.Enqueue(ctx, job))
	time.Sleep(80 * time.Millisecond)
	require.EqualValues(t, 1, atomic.LoadInt32(&calls))
}

func TestEnqueueCanceledWhenBufferFull(t *testing.T) {
	p := NewPool(1, 2, 2)
	ctx := context.Background()
	noop := func(context.Context) {}

	require.NoError(t, p.Enqueue(ctx, noop))
	require.NoError(t, p.Enqueue(ctx, noop))

	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel()

	err := p.Enqueue(ctxCancel, noop)
	require.ErrorIs(t, err, context.Canceled)
}

func TestRunWaitsForContext(t *testing.T) {
	p := NewPool(1, 2, 4)
	ctx, cancel := context.WithCancel(context.Background())

	runDone := make(chan error, 1)
	go func() { runDone <- p.Run(ctx) }()

	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case err := <-runDone:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return")
	}
}

func TestRunExecutesJobs(t *testing.T) {
	p := NewPool(2, 4, 32)
	ctx, cancel := context.WithCancel(context.Background())

	runErr := make(chan error, 1)
	go func() { runErr <- p.Run(ctx) }()

	time.Sleep(40 * time.Millisecond)

	const n = 10
	var ran int32
	for i := 0; i < n; i++ {
		require.NoError(t, p.Enqueue(ctx, func(context.Context) {
			atomic.AddInt32(&ran, 1)
		}))
	}

	waitUntil(t, func() bool { return atomic.LoadInt32(&ran) == n }, 2*time.Second, "jobs not all executed")
	cancel()

	select {
	case err := <-runErr:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not finish")
	}
}

func TestRemoveWorkerRespectsMin(t *testing.T) {
	p := NewPool(3, 5, 10)
	p.stops = make([]chan struct{}, 3)
	for i := range p.stops {
		p.stops[i] = make(chan struct{})
	}
	p.removeWorker()
	require.Len(t, p.stops, 3, "at min workers, remove is a no-op")
}

func TestRemoveWorkerShrinksWhenAboveMin(t *testing.T) {
	p := NewPool(2, 5, 10)
	ch1, ch2, ch3 := make(chan struct{}), make(chan struct{}), make(chan struct{})
	p.stops = []chan struct{}{ch1, ch2, ch3}

	p.removeWorker()
	require.Len(t, p.stops, 2)

	select {
	case <-ch3:
	default:
		t.Fatal("removed worker stop channel should be closed")
	}
}

func TestScaleUpPastUpAt(t *testing.T) {
	if testing.Short() {
		t.Skip("timing-dependent scale test")
	}

	p := NewPool(2, 6, 24)
	ctx, cancel := context.WithCancel(context.Background())

	runErr := make(chan error, 1)
	go func() { runErr <- p.Run(ctx) }()

	time.Sleep(30 * time.Millisecond)

	block := make(chan struct{})
	blockingJob := func(context.Context) { <-block }

	require.NoError(t, p.Enqueue(ctx, blockingJob))
	require.NoError(t, p.Enqueue(ctx, blockingJob))

	for i := 0; i < 17; i++ {
		require.NoError(t, p.Enqueue(ctx, func(context.Context) {}))
	}

	// len(inputCh) > upAt (18), next tick should add workers
	time.Sleep(250 * time.Millisecond)
	p.mu.Lock()
	nWorkers := len(p.stops)
	p.mu.Unlock()
	require.Greater(t, nWorkers, 2, "expected scale up when queue length exceeds upAt")

	close(block)
	cancel()

	select {
	case err := <-runErr:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not finish")
	}
}

func waitUntil(t *testing.T, cond func() bool, timeout time.Duration, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal(msg)
}
