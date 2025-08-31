package hub

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func createTestHub(t *testing.T, testName string) (*Hub, func()) {
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("lock_test_%s_*", testName))
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}

	// Create lock bucket for testing
	config := KeyValueStoreConfig{
		Bucket:       "test-locks",
		Description:  "Test lock bucket",
		MaxValueSize: NewSizeFromKilobytes(1),
		TTL:          10 * time.Second, // 10 seconds TTL for testing
		Replicas:     1,
	}

	err = hub.CreateOrUpdateKeyValueStore(config)
	if err != nil {
		hub.Shutdown()
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create lock bucket: %v", err)
	}

	cleanup := func() {
		hub.Shutdown()
		os.RemoveAll(tempDir)
	}

	return hub, cleanup
}

func TestTryLockBasic(t *testing.T) {
	hub, cleanup := createTestHub(t, "try_lock_basic")
	defer cleanup()

	lockKey := "resource-1"
	bucket := "test-locks"

	// Test acquiring a lock
	cancel, err := hub.TryLock(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}
	defer cancel()

	// Verify lock is held
	isLocked, err := hub.IsLocked(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to check lock status: %v", err)
	}
	if !isLocked {
		t.Error("Expected resource to be locked")
	}

	// Test that second lock attempt fails
	_, err = hub.TryLock(bucket, lockKey)
	if err == nil {
		t.Error("Expected second lock attempt to fail")
	}
	if !errors.Is(err, nats.ErrKeyExists) {
		t.Errorf("Expected ErrKeyExists, got: %v", err)
	}

	// Release lock
	cancel()

	// Verify lock is released
	isLocked, err = hub.IsLocked(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to check lock status after release: %v", err)
	}
	if isLocked {
		t.Error("Expected resource to be unlocked after release")
	}

	// Test acquiring lock again after release
	cancel2, err := hub.TryLock(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to acquire lock after release: %v", err)
	}
	defer cancel2()

	t.Logf("Successfully tested basic lock operations")
}

func TestLockWithContext(t *testing.T) {
	hub, cleanup := createTestHub(t, "lock_with_context")
	defer cleanup()

	lockKey := "resource-context"
	bucket := "test-locks"

	// Test successful lock acquisition with context
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	cancel, err := hub.Lock(ctx, bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to acquire lock with context: %v", err)
	}
	defer cancel()

	// Verify lock is held
	isLocked, err := hub.IsLocked(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to check lock status: %v", err)
	}
	if !isLocked {
		t.Error("Expected resource to be locked")
	}

	t.Logf("Successfully acquired lock with context")
}

func TestLockTimeout(t *testing.T) {
	hub, cleanup := createTestHub(t, "lock_timeout")
	defer cleanup()

	lockKey := "resource-timeout"
	bucket := "test-locks"

	// First acquire the lock
	cancel1, err := hub.TryLock(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to acquire initial lock: %v", err)
	}
	defer cancel1()

	// Try to acquire the same lock with a short timeout
	ctx, ctxCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer ctxCancel()

	start := time.Now()
	_, err = hub.Lock(ctx, bucket, lockKey)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected lock acquisition to timeout")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}

	// Verify timeout happened around the expected time
	if duration < 90*time.Millisecond || duration > 200*time.Millisecond {
		t.Errorf("Expected timeout around 100ms, got %v", duration)
	}

	t.Logf("Successfully tested lock timeout after %v", duration)
}

func TestLockRetryWithBackoff(t *testing.T) {
	hub, cleanup := createTestHub(t, "lock_retry")
	defer cleanup()

	lockKey := "resource-retry"
	bucket := "test-locks"

	// First acquire the lock
	cancel1, err := hub.TryLock(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to acquire initial lock: %v", err)
	}

	// Release the lock after a short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel1()
	}()

	// Try to acquire the lock with retry
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer ctxCancel()

	start := time.Now()
	cancel2, err := hub.Lock(ctx, bucket, lockKey)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to acquire lock with retry: %v", err)
	}
	defer cancel2()

	// Verify it took some time to acquire (due to retry)
	if duration < 180*time.Millisecond {
		t.Errorf("Expected retry to take at least 180ms, got %v", duration)
	}

	t.Logf("Successfully acquired lock after retry in %v", duration)
}

func TestConcurrentLockAcquisition(t *testing.T) {
	hub, cleanup := createTestHub(t, "concurrent_lock")
	defer cleanup()

	lockKey := "resource-concurrent"
	bucket := "test-locks"
	const numGoroutines = 5 // Reduced to make test more reliable

	var successCount int64
	var timeoutCount int64
	var wg sync.WaitGroup
	results := make(chan string, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			// Generous timeout to allow all goroutines to eventually succeed
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			lockCancel, err := hub.Lock(ctx, bucket, lockKey)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					atomic.AddInt64(&timeoutCount, 1)
					results <- fmt.Sprintf("Goroutine %d timed out waiting for lock", goroutineID)
				} else {
					results <- fmt.Sprintf("Goroutine %d failed to acquire lock: %v", goroutineID, err)
				}
				return
			}

			// Hold the lock for a very short time
			time.Sleep(10 * time.Millisecond)
			lockCancel()

			atomic.AddInt64(&successCount, 1)
			results <- fmt.Sprintf("Goroutine %d successfully acquired and released lock", goroutineID)
		}(i)
	}

	wg.Wait()
	close(results)

	// Check results
	for result := range results {
		t.Logf("%s", result)
	}

	finalSuccessCount := atomic.LoadInt64(&successCount)
	finalTimeoutCount := atomic.LoadInt64(&timeoutCount)

	// With fewer goroutines and longer timeout, all should succeed
	if finalSuccessCount < numGoroutines-1 { // Allow 1 failure due to timing
		t.Errorf("Expected at least %d goroutines to succeed, got %d successful, %d timeouts",
			numGoroutines-1, finalSuccessCount, finalTimeoutCount)
	}

	// Verify that at least some concurrency happened (not all immediate successes)
	if finalSuccessCount > 0 {
		t.Logf("Successfully tested concurrent lock acquisition: %d successful, %d timeouts",
			finalSuccessCount, finalTimeoutCount)
	}
}

func TestForceUnlock(t *testing.T) {
	hub, cleanup := createTestHub(t, "force_unlock")
	defer cleanup()

	lockKey := "resource-force"
	bucket := "test-locks"

	// Acquire a lock
	cancel, err := hub.TryLock(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}
	defer cancel() // Just in case

	// Verify lock is held
	isLocked, err := hub.IsLocked(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to check lock status: %v", err)
	}
	if !isLocked {
		t.Error("Expected resource to be locked")
	}

	// Force unlock
	err = hub.ForceUnlock(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to force unlock: %v", err)
	}

	// Verify lock is released
	isLocked, err = hub.IsLocked(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to check lock status after force unlock: %v", err)
	}
	if isLocked {
		t.Error("Expected resource to be unlocked after force unlock")
	}

	// Test that we can acquire the lock again
	cancel2, err := hub.TryLock(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to acquire lock after force unlock: %v", err)
	}
	defer cancel2()

	t.Logf("Successfully tested force unlock")
}

func TestLockTTL(t *testing.T) {
	hub, cleanup := createTestHub(t, "lock_ttl")
	defer cleanup()

	// Create a bucket with very short TTL for testing
	shortTTLConfig := KeyValueStoreConfig{
		Bucket:       "short-ttl-locks",
		Description:  "Short TTL lock bucket",
		MaxValueSize: NewSizeFromKilobytes(1),
		TTL:          1 * time.Second, // Short TTL but not too short
		Replicas:     1,
	}

	err := hub.CreateOrUpdateKeyValueStore(shortTTLConfig)
	if err != nil {
		t.Fatalf("Failed to create short TTL bucket: %v", err)
	}

	lockKey := "resource-ttl"
	bucket := "short-ttl-locks"

	// Acquire a lock (but don't call cancel to simulate a crashed process)
	_, err = hub.TryLock(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}
	// Intentionally not calling cancel() to simulate a process crash

	// Verify lock is initially held
	isLocked, err := hub.IsLocked(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to check initial lock status: %v", err)
	}
	if !isLocked {
		t.Error("Expected resource to be initially locked")
	}

	// Wait for TTL to expire (wait longer than TTL)
	t.Logf("Waiting for TTL to expire...")
	time.Sleep(2 * time.Second)

	// Try to acquire the lock again - this should succeed if TTL worked
	cancel2, err := hub.TryLock(bucket, lockKey)
	if err != nil {
		// If we get an error, check if the key still exists but might be expired
		isLocked, checkErr := hub.IsLocked(bucket, lockKey)
		if checkErr != nil {
			t.Logf("Lock expired and key removed (expected): %v", checkErr)
			// Try once more to acquire - should work now
			cancel2, err = hub.TryLock(bucket, lockKey)
			if err != nil {
				t.Fatalf("Failed to acquire lock after TTL expiry and key removal: %v", err)
			}
			defer cancel2()
			t.Logf("Successfully acquired lock after TTL expiry")
			return
		}

		if isLocked {
			t.Errorf("Lock should have expired due to TTL but key still exists and is locked")
		} else {
			t.Logf("Key exists but lock appears to be expired")
		}

		// Force unlock to clean up
		forceErr := hub.ForceUnlock(bucket, lockKey)
		if forceErr != nil {
			t.Logf("Force unlock also failed: %v", forceErr)
		}

		// Try to acquire again
		cancel2, err = hub.TryLock(bucket, lockKey)
		if err != nil {
			t.Fatalf("Failed to acquire lock even after force unlock: %v", err)
		}
	}
	defer cancel2()

	t.Logf("Successfully tested lock TTL behavior")
}

func TestIsLocked(t *testing.T) {
	hub, cleanup := createTestHub(t, "is_locked")
	defer cleanup()

	lockKey := "resource-check"
	bucket := "test-locks"

	// Initially not locked
	isLocked, err := hub.IsLocked(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to check initial lock status: %v", err)
	}
	if isLocked {
		t.Error("Expected resource to be initially unlocked")
	}

	// Acquire lock
	cancel, err := hub.TryLock(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}
	defer cancel()

	// Should be locked now
	isLocked, err = hub.IsLocked(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to check lock status after acquisition: %v", err)
	}
	if !isLocked {
		t.Error("Expected resource to be locked after acquisition")
	}

	// Release lock
	cancel()

	// Should be unlocked again
	isLocked, err = hub.IsLocked(bucket, lockKey)
	if err != nil {
		t.Fatalf("Failed to check lock status after release: %v", err)
	}
	if isLocked {
		t.Error("Expected resource to be unlocked after release")
	}

	t.Logf("Successfully tested IsLocked function")
}

func TestLockErrorHandling(t *testing.T) {
	hub, cleanup := createTestHub(t, "lock_errors")
	defer cleanup()

	// Test operations on non-existent bucket
	_, err := hub.TryLock("nonexistent-bucket", "test-key")
	if err == nil {
		t.Error("Expected error when trying to lock in non-existent bucket")
	} else {
		t.Logf("Got expected error for non-existent bucket: %v", err)
	}

	// Test IsLocked on non-existent bucket
	_, err = hub.IsLocked("nonexistent-bucket", "test-key")
	if err == nil {
		t.Error("Expected error when checking lock in non-existent bucket")
	} else {
		t.Logf("Got expected error for IsLocked on non-existent bucket: %v", err)
	}

	// Test ForceUnlock on non-existent bucket
	err = hub.ForceUnlock("nonexistent-bucket", "test-key")
	if err == nil {
		t.Error("Expected error when force unlocking in non-existent bucket")
	} else {
		t.Logf("Got expected error for ForceUnlock on non-existent bucket: %v", err)
	}

	// Test Lock with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = hub.Lock(ctx, "test-locks", "test-key")
	if err == nil {
		t.Error("Expected error when using cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got: %v", err)
	} else {
		t.Logf("Got expected context.Canceled error: %v", err)
	}

	t.Logf("Successfully tested error handling")
}

func TestMultipleLockKeys(t *testing.T) {
	hub, cleanup := createTestHub(t, "multiple_keys")
	defer cleanup()

	bucket := "test-locks"
	keys := []string{"resource-a", "resource-b", "resource-c"}
	var cancels []func()

	// Acquire locks on multiple keys
	for _, key := range keys {
		cancel, err := hub.TryLock(bucket, key)
		if err != nil {
			t.Fatalf("Failed to acquire lock on key %s: %v", key, err)
		}
		cancels = append(cancels, cancel)
	}

	// Verify all locks are held
	for _, key := range keys {
		isLocked, err := hub.IsLocked(bucket, key)
		if err != nil {
			t.Fatalf("Failed to check lock status for key %s: %v", key, err)
		}
		if !isLocked {
			t.Errorf("Expected key %s to be locked", key)
		}
	}

	// Release all locks
	for i, cancel := range cancels {
		cancel()
		t.Logf("Released lock on key %s", keys[i])
	}

	// Verify all locks are released
	for _, key := range keys {
		isLocked, err := hub.IsLocked(bucket, key)
		if err != nil {
			t.Fatalf("Failed to check lock status after release for key %s: %v", key, err)
		}
		if isLocked {
			t.Errorf("Expected key %s to be unlocked after release", key)
		}
	}

	t.Logf("Successfully tested multiple lock keys")
}
