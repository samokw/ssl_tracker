package ssl

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

// TestMain runs after all tests and checks for goroutine leaks.
// If any goroutines are still running that shouldn't be, the test fails.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// drainResults is a helper that reads all results from the pool.
// Without this, the pool would block trying to send results.
func drainResults(wp *WorkerPool) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		for range wp.GetResults() {
		}
		close(done)
	}()
	return done
}

// TestWorkerPool_Basic - the simplest test: add one task, get one result.
func TestWorkerPool_Basic(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Create a pool with 1 worker
	wp := NewWorkerPool(1)
	wp.Start()

	// Add a task
	wp.AddTask(Task{Domain: "example.com", DomainID: 1, UserID: 1})

	// Get the result
	result := <-wp.GetResults()

	// Check that the result matches our task
	assert.Equal(t, "example.com", result.Task.Domain)
	assert.Equal(t, 1, result.Task.DomainID)

	drainResults(wp)
	wp.Stop()
}

// TestWorkerPool_InvalidDomain - bad domains should return errors, not crash.
func TestWorkerPool_InvalidDomain(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := NewWorkerPool(1)
	wp.Start()

	// Empty domain is invalid
	wp.AddTask(Task{Domain: "", DomainID: 1, UserID: 1})

	result := <-wp.GetResults()

	// Should have an error
	assert.Error(t, result.Error)
	assert.Nil(t, result.Certificate)

	drainResults(wp)
	wp.Stop()
}

// TestWorkerPool_StopsCleanly - Stop() should finish quickly, not hang.
func TestWorkerPool_StopsCleanly(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := NewWorkerPool(5)
	wp.Start()

	wp.AddTask(Task{Domain: "example.com", DomainID: 1, UserID: 1})

	done := drainResults(wp)
	time.Sleep(50 * time.Millisecond)

	// Run Stop() in a goroutine so we can timeout if it hangs
	stopped := make(chan struct{})
	go func() {
		wp.Stop()
		close(stopped)
	}()

	// Stop should finish within 2 seconds
	select {
	case <-stopped:
		// Good!
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() took too long - might be a deadlock")
	}

	<-done
}

// TestWorkerPool_ConcurrentAdds - adding tasks from many goroutines at once is safe.
func TestWorkerPool_ConcurrentAdds(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := NewWorkerPool(5)
	wp.Start()

	// Count results
	var count atomic.Int32
	done := make(chan struct{})
	go func() {
		for range wp.GetResults() {
			count.Add(1)
		}
		close(done)
	}()

	// 10 goroutines each add 10 tasks = 100 total
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				wp.AddTask(Task{
					Domain:   "test.com",
					DomainID: id*10 + j,
					UserID:   id,
				})
			}
		}(i)
	}

	wg.Wait()
	wp.Stop()
	<-done

	// All 100 tasks should be processed
	assert.Equal(t, int32(100), count.Load())
}

// TestWorkerPool_ResultHasTimestamp - each result should have a CheckedAt time.
func TestWorkerPool_ResultHasTimestamp(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := NewWorkerPool(1)
	wp.Start()

	wp.AddTask(Task{Domain: "example.com", DomainID: 1, UserID: 1})

	result := <-wp.GetResults()

	// CheckedAt should not be zero
	assert.False(t, result.CheckedAt.IsZero())

	drainResults(wp)
	wp.Stop()
}

// TestWorkerPool_ZeroWorkers - edge case: pool with 0 workers should still stop.
func TestWorkerPool_ZeroWorkers(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := NewWorkerPool(0)
	wp.Start()

	wp.AddTask(Task{Domain: "example.com", DomainID: 1, UserID: 1})

	// Stop should still work even with no workers
	stopped := make(chan struct{})
	go func() {
		wp.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
		// Good
	case <-time.After(time.Second):
		t.Fatal("Stop() blocked with zero workers")
	}
}

// TestWorkerPool_HighLoad - stress test with lots of tasks (skipped in short mode).
func TestWorkerPool_HighLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test")
	}

	defer goleak.VerifyNone(t)

	wp := NewWorkerPool(50)
	wp.Start()

	var count atomic.Int32
	done := make(chan struct{})
	go func() {
		for range wp.GetResults() {
			count.Add(1)
		}
		close(done)
	}()

	// Add 500 tasks
	for i := 0; i < 500; i++ {
		wp.AddTask(Task{Domain: "test.com", DomainID: i, UserID: 1})
	}

	time.Sleep(200 * time.Millisecond)
	wp.Stop()
	<-done

	assert.Equal(t, int32(500), count.Load())
}
