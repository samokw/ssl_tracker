package ssl

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

// TestCertService_Basic - start the service, check a domain, stop it.
func TestCertService_Basic(t *testing.T) {
	defer goleak.VerifyNone(t)

	cs := NewCertService()
	cs.Start()

	cs.CheckDomain("example.com", 1, 1)

	time.Sleep(100 * time.Millisecond)
	cs.Stop()
}

// TestCertService_ResultHandler - set a handler to receive results.
func TestCertService_ResultHandler(t *testing.T) {
	defer goleak.VerifyNone(t)

	cs := NewCertService()

	var gotResult atomic.Bool
	cs.SetResultHandler(func(r Result) {
		gotResult.Store(true)
	})

	cs.Start()
	cs.CheckDomain("invalid..domain", 1, 1) // invalid domain = quick error

	time.Sleep(200 * time.Millisecond)
	cs.Stop()

	assert.True(t, gotResult.Load(), "Handler should have been called")
}

// TestCertService_MultipleStartsOK - calling Start() multiple times is fine.
func TestCertService_MultipleStartsOK(t *testing.T) {
	defer goleak.VerifyNone(t)

	cs := NewCertService()

	// These extra calls should be ignored
	cs.Start()
	cs.Start()
	cs.Start()

	cs.CheckDomain("example.com", 1, 1)

	time.Sleep(50 * time.Millisecond)
	cs.Stop()
}

// TestCertService_StopWithoutStart - Stop() shouldn't hang if never started.
func TestCertService_StopWithoutStart(t *testing.T) {
	defer goleak.VerifyNone(t)

	cs := NewCertService()

	// Run Stop in a goroutine so we can timeout if it hangs
	done := make(chan struct{})
	go func() {
		cs.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Good!
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() hung without Start()")
	}
}

// TestCertService_NoHandler - works even without setting a handler.
func TestCertService_NoHandler(t *testing.T) {
	defer goleak.VerifyNone(t)

	cs := NewCertService()
	// Don't set a handler - should use default

	cs.Start()
	cs.CheckDomain("invalid..domain", 1, 1)

	time.Sleep(100 * time.Millisecond)
	cs.Stop()
}

// TestCertService_HighLoad - stress test with many domains (skipped in short mode).
func TestCertService_HighLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test")
	}

	defer goleak.VerifyNone(t)

	cs := NewCertService()

	var count atomic.Int32
	cs.SetResultHandler(func(r Result) {
		count.Add(1)
	})

	cs.Start()

	// Check 100 domains
	for i := 0; i < 100; i++ {
		cs.CheckDomain("invalid..test", i, 1)
	}

	time.Sleep(500 * time.Millisecond)
	cs.Stop()

	assert.Equal(t, int32(100), count.Load())
}
