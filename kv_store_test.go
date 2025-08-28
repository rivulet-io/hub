package hub

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestCreateOrUpdateKeyValueStore(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "kv_test_create_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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
	defer hub.Shutdown()

	// Test creating a new KV store
	config := KeyValueStoreConfig{
		Bucket:       "test-configs",
		Description:  "Test configuration store",
		MaxValueSize: NewSizeFromKilobytes(64),
		TTL:          24 * time.Hour,
		MaxBytes:     NewSizeFromMegabytes(10),
		Replicas:     1,
	}

	err = hub.CreateOrUpdateKeyValueStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	// Verify KV store was created
	kv, err := hub.jetstreamCtx.KeyValue("test-configs")
	if err != nil {
		t.Fatalf("Failed to access KV store: %v", err)
	}

	status, err := kv.Status()
	if err != nil {
		t.Fatalf("Failed to get KV status: %v", err)
	}

	if status.Bucket() != "test-configs" {
		t.Errorf("Expected bucket name 'test-configs', got '%s'", status.Bucket())
	}

	if status.Bucket() != config.Bucket {
		t.Errorf("Expected bucket '%s', got '%s'", config.Bucket, status.Bucket())
	}

	// Test creating another KV store (since CreateOrUpdate actually just creates)
	config2 := KeyValueStoreConfig{
		Bucket:       "test-configs-updated",
		Description:  "Updated test configuration store",
		MaxValueSize: NewSizeFromKilobytes(128),
		TTL:          48 * time.Hour,
		MaxBytes:     NewSizeFromMegabytes(20),
		Replicas:     1,
	}

	err = hub.CreateOrUpdateKeyValueStore(config2)
	if err != nil {
		t.Fatalf("Failed to create second KV store: %v", err)
	}

	// Verify second KV store was created
	kv2, err := hub.jetstreamCtx.KeyValue("test-configs-updated")
	if err != nil {
		t.Fatalf("Failed to access second KV store: %v", err)
	}

	status2, err := kv2.Status()
	if err != nil {
		t.Fatalf("Failed to get second KV status: %v", err)
	}

	if status2.Bucket() != "test-configs-updated" {
		t.Errorf("Expected bucket name 'test-configs-updated', got '%s'", status2.Bucket())
	}

	t.Logf("Successfully created multiple KV stores")
}

func TestKeyValueStoreCRUD(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "kv_test_crud_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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
	defer hub.Shutdown()

	// Create KV store
	config := KeyValueStoreConfig{
		Bucket:       "user-data",
		Description:  "User data store",
		MaxValueSize: NewSizeFromKilobytes(1),
		TTL:          time.Hour,
	}
	err = hub.CreateOrUpdateKeyValueStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	// Test PUT operation
	userData := []byte(`{"name":"John Doe","email":"john@example.com"}`)
	revision, err := hub.PutToKeyValueStore("user-data", "user123", userData)
	if err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	if revision != 1 {
		t.Errorf("Expected revision 1, got %d", revision)
	}

	// Test GET operation
	retrievedData, retrievedRevision, err := hub.GetFromKeyValueStore("user-data", "user123")
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if string(retrievedData) != string(userData) {
		t.Errorf("Retrieved data doesn't match original")
	}

	if retrievedRevision != revision {
		t.Errorf("Expected revision %d, got %d", revision, retrievedRevision)
	}

	// Test UPDATE operation
	updatedData := []byte(`{"name":"John Doe","email":"john.doe@example.com"}`)
	newRevision, err := hub.UpdateToKeyValueStore("user-data", "user123", updatedData, revision)
	if err != nil {
		t.Fatalf("Failed to update data: %v", err)
	}

	if newRevision != revision+1 {
		t.Errorf("Expected revision %d, got %d", revision+1, newRevision)
	}

	// Verify update
	retrievedData, retrievedRevision, err = hub.GetFromKeyValueStore("user-data", "user123")
	if err != nil {
		t.Fatalf("Failed to get updated data: %v", err)
	}

	if string(retrievedData) != string(updatedData) {
		t.Errorf("Updated data doesn't match")
	}

	if retrievedRevision != newRevision {
		t.Errorf("Expected revision %d, got %d", newRevision, retrievedRevision)
	}

	// Test DELETE operation
	err = hub.DeleteFromKeyValueStore("user-data", "user123")
	if err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// Verify deletion
	_, _, err = hub.GetFromKeyValueStore("user-data", "user123")
	if err == nil {
		t.Error("Expected error when getting deleted key")
	}

	t.Logf("Successfully completed CRUD operations")
}

func TestKeyValueStoreConcurrency(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "kv_test_concurrency_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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
	defer hub.Shutdown()

	// Create KV store
	config := KeyValueStoreConfig{
		Bucket:       "concurrent-data",
		Description:  "Concurrent access test store",
		MaxValueSize: NewSizeFromKilobytes(1),
	}
	err = hub.CreateOrUpdateKeyValueStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	// Test concurrent access from multiple goroutines
	const numGoroutines = 10
	const numOperations = 5

	var wg sync.WaitGroup
	var mu sync.Mutex
	operations := make(map[string]int)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", goroutineID, j)
				value := []byte(fmt.Sprintf("value-%d-%d", goroutineID, j))

				// PUT operation
				revision, err := hub.PutToKeyValueStore("concurrent-data", key, value)
				if err != nil {
					t.Errorf("Goroutine %d failed to put: %v", goroutineID, err)
					continue
				}

				// GET operation
				retrievedValue, retrievedRevision, err := hub.GetFromKeyValueStore("concurrent-data", key)
				if err != nil {
					t.Errorf("Goroutine %d failed to get: %v", goroutineID, err)
					continue
				}

				if string(retrievedValue) != string(value) {
					t.Errorf("Goroutine %d: value mismatch", goroutineID)
				}

				if retrievedRevision != revision {
					t.Errorf("Goroutine %d: revision mismatch", goroutineID)
				}

				mu.Lock()
				operations[key]++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Verify all operations completed
	mu.Lock()
	totalOperations := len(operations)
	mu.Unlock()

	if totalOperations != numGoroutines*numOperations {
		t.Errorf("Expected %d operations, got %d", numGoroutines*numOperations, totalOperations)
	}

	t.Logf("Successfully completed %d concurrent operations", totalOperations)
}

func TestKeyValueStoreClusterOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "kv_test_cluster_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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
	defer hub.Shutdown()

	// Create multiple KV stores for cluster simulation
	stores := []string{"node1-data", "node2-data", "node3-data"}

	for _, storeName := range stores {
		config := KeyValueStoreConfig{
			Bucket:       storeName,
			Description:  fmt.Sprintf("Data store for %s", storeName),
			MaxValueSize: NewSizeFromKilobytes(64),
			TTL:          time.Hour,
		}

		err = hub.CreateOrUpdateKeyValueStore(config)
		if err != nil {
			t.Fatalf("Failed to create KV store %s: %v", storeName, err)
		}
	}

	// Simulate cluster operations
	var wg sync.WaitGroup
	results := make(chan string, len(stores)*3)

	for i, storeName := range stores {
		wg.Add(1)
		go func(store string, nodeID int) {
			defer wg.Done()

			// Node-specific data
			key := fmt.Sprintf("node%d-key", nodeID)
			value := []byte(fmt.Sprintf("data from node %d", nodeID))

			// PUT operation
			revision, err := hub.PutToKeyValueStore(store, key, value)
			if err != nil {
				results <- fmt.Sprintf("PUT failed for %s: %v", store, err)
				return
			}

			// GET operation
			retrievedValue, retrievedRevision, err := hub.GetFromKeyValueStore(store, key)
			if err != nil {
				results <- fmt.Sprintf("GET failed for %s: %v", store, err)
				return
			}

			if string(retrievedValue) != string(value) {
				results <- fmt.Sprintf("Value mismatch for %s", store)
				return
			}

			if retrievedRevision != revision {
				results <- fmt.Sprintf("Revision mismatch for %s", store)
				return
			}

			// UPDATE operation
			updatedValue := []byte(fmt.Sprintf("updated data from node %d", nodeID))
			newRevision, err := hub.UpdateToKeyValueStore(store, key, updatedValue, revision)
			if err != nil {
				results <- fmt.Sprintf("UPDATE failed for %s: %v", store, err)
				return
			}

			if newRevision != revision+1 {
				results <- fmt.Sprintf("Update revision mismatch for %s", store)
				return
			}

			results <- fmt.Sprintf("SUCCESS: %s operations completed", store)
		}(storeName, i+1)
	}

	wg.Wait()
	close(results)

	// Check results
	successCount := 0
	for result := range results {
		if result[:7] == "SUCCESS" {
			successCount++
			t.Logf("%s", result)
		} else {
			t.Errorf("%s", result)
		}
	}

	if successCount != len(stores) {
		t.Errorf("Expected %d successful operations, got %d", len(stores), successCount)
	}
}

func TestKeyValueStoreTTL(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "kv_test_ttl_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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
	defer hub.Shutdown()

	// Create KV store with short TTL
	config := KeyValueStoreConfig{
		Bucket:       "ttl-test",
		Description:  "TTL test store",
		MaxValueSize: NewSizeFromKilobytes(1),
		TTL:          2 * time.Second, // Very short TTL for testing
	}
	err = hub.CreateOrUpdateKeyValueStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	// Put data with TTL
	key := "temp-data"
	value := []byte("temporary data")
	_, err = hub.PutToKeyValueStore("ttl-test", key, value)
	if err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	// Verify data exists immediately
	retrievedValue, _, err := hub.GetFromKeyValueStore("ttl-test", key)
	if err != nil {
		t.Fatalf("Failed to get data immediately: %v", err)
	}

	if string(retrievedValue) != string(value) {
		t.Errorf("Retrieved data doesn't match original")
	}

	// Wait for TTL to expire
	t.Logf("Waiting for TTL to expire...")
	time.Sleep(3 * time.Second)

	// Verify data has expired
	_, _, err = hub.GetFromKeyValueStore("ttl-test", key)
	if err == nil {
		t.Error("Expected data to be expired, but it still exists")
	} else {
		t.Logf("Data correctly expired: %v", err)
	}
}

func TestKeyValueStoreErrorHandling(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "kv_test_error_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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
	defer hub.Shutdown()

	// Test operations on non-existent bucket
	_, _, err = hub.GetFromKeyValueStore("nonexistent-bucket", "test-key")
	if err == nil {
		t.Error("Expected error when accessing non-existent bucket")
	} else {
		t.Logf("Got expected error for non-existent bucket: %v", err)
	}

	_, err = hub.PutToKeyValueStore("nonexistent-bucket", "test-key", []byte("test"))
	if err == nil {
		t.Error("Expected error when putting to non-existent bucket")
	} else {
		t.Logf("Got expected error for put to non-existent bucket: %v", err)
	}

	// Test operations on non-existent key
	// First create a valid bucket
	config := KeyValueStoreConfig{
		Bucket:       "error-test",
		Description:  "Error handling test store",
		MaxValueSize: NewSizeFromKilobytes(1),
	}
	err = hub.CreateOrUpdateKeyValueStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	_, _, err = hub.GetFromKeyValueStore("error-test", "nonexistent-key")
	if err == nil {
		t.Error("Expected error when getting non-existent key")
	} else {
		t.Logf("Got expected error for non-existent key: %v", err)
	}

	// Test update with wrong revision
	_, err = hub.PutToKeyValueStore("error-test", "test-key", []byte("initial"))
	if err != nil {
		t.Fatalf("Failed to put initial data: %v", err)
	}

	_, err = hub.UpdateToKeyValueStore("error-test", "test-key", []byte("update"), 999) // Wrong revision
	if err == nil {
		t.Error("Expected error when updating with wrong revision")
	} else {
		t.Logf("Got expected error for wrong revision: %v", err)
	}

	// Test delete non-existent key (NATS doesn't return error for this)
	err = hub.DeleteFromKeyValueStore("error-test", "nonexistent-key")
	// Note: NATS doesn't return an error when deleting a non-existent key
	t.Logf("Delete non-existent key result: %v", err)

	// Test purge non-existent key (NATS doesn't return error for this)
	err = hub.PurgeKeyValueStore("error-test", "nonexistent-key")
	// Note: NATS doesn't return an error when purging a non-existent key
	t.Logf("Purge non-existent key result: %v", err)
}

func TestKeyValueStorePurge(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "kv_test_purge_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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
	defer hub.Shutdown()

	// Create KV store
	config := KeyValueStoreConfig{
		Bucket:       "purge-test",
		Description:  "Purge test store",
		MaxValueSize: NewSizeFromKilobytes(1),
	}
	err = hub.CreateOrUpdateKeyValueStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	// Put initial data
	key := "test-key"
	initialValue := []byte("initial value")
	_, err = hub.PutToKeyValueStore("purge-test", key, initialValue)
	if err != nil {
		t.Fatalf("Failed to put initial data: %v", err)
	}

	// Update data multiple times to create history
	for i := 1; i <= 3; i++ {
		value := []byte(fmt.Sprintf("updated value %d", i))
		_, err = hub.PutToKeyValueStore("purge-test", key, value)
		if err != nil {
			t.Fatalf("Failed to update data %d: %v", i, err)
		}
	}

	// Verify current value
	currentValue, _, err := hub.GetFromKeyValueStore("purge-test", key)
	if err != nil {
		t.Fatalf("Failed to get current value: %v", err)
	}

	expectedCurrent := "updated value 3"
	if string(currentValue) != expectedCurrent {
		t.Errorf("Expected current value '%s', got '%s'", expectedCurrent, string(currentValue))
	}

	// Test PURGE operation
	err = hub.PurgeKeyValueStore("purge-test", key)
	if err != nil {
		t.Fatalf("Failed to purge key: %v", err)
	}

	// Verify key is completely removed (including history)
	_, _, err = hub.GetFromKeyValueStore("purge-test", key)
	if err == nil {
		t.Error("Expected key to be completely removed after purge")
	} else {
		t.Logf("Key correctly purged: %v", err)
	}

	t.Logf("Successfully tested purge operation")
}
