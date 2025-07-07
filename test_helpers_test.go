//go:build integration

package sendria_test

import (
	"testing"
	"time"
)

// waitFor retries a condition function until it returns true or timeout occurs.
// This is more idiomatic than using hard-coded sleeps.
func waitFor(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration) {
	t.Helper()
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}
	t.Fatalf("condition not met within %v", timeout)
}

// eventually is similar to waitFor but doesn't fail the test, just returns whether condition was met
func eventually(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration) bool {
	t.Helper()
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}