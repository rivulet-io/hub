package hub

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestCreateObjectStore(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "obj_test_create_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
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
	defer hub.Shutdown()

	// Test creating a new object store
	config := ObjectStoreConfig{
		Bucket:      "test-objects",
		Description: "Test object store",
		TTL:         24 * time.Hour,
		MaxBytes:    NewSizeFromMegabytes(100),
		Replicas:    1,
		Metadata: map[string]string{
			"environment": "test",
			"version":     "1.0",
		},
	}

	err = hub.CreateObjectStore(config)
	if err != nil {
		t.Fatalf("Failed to create object store: %v", err)
	}

	// Verify object store was created by trying to access it
	store, err := hub.jetstreamCtx.ObjectStore("test-objects")
	if err != nil {
		t.Fatalf("Failed to access object store: %v", err)
	}

	// Try to get status (this will verify the store exists)
	_, err = store.Status()
	if err != nil {
		t.Fatalf("Failed to get object store status: %v", err)
	}

	// Test creating another object store
	config2 := ObjectStoreConfig{
		Bucket:      "test-objects-2",
		Description: "Second test object store",
		TTL:         48 * time.Hour,
		MaxBytes:    NewSizeFromMegabytes(200),
		Replicas:    1,
	}

	err = hub.CreateObjectStore(config2)
	if err != nil {
		t.Fatalf("Failed to create second object store: %v", err)
	}

	t.Logf("Successfully created multiple object stores")
}

func TestObjectStoreCRUD(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "obj_test_crud_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
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
	defer hub.Shutdown()

	// Create object store
	config := ObjectStoreConfig{
		Bucket:      "documents",
		Description: "Document storage",
		TTL:         time.Hour,
		MaxBytes:    NewSizeFromMegabytes(50),
		Replicas:    1,
	}
	err = hub.CreateObjectStore(config)
	if err != nil {
		t.Fatalf("Failed to create object store: %v", err)
	}

	// Test PUT operation
	documentData := []byte(`{"title":"Test Document","content":"This is a test document","author":"Test Author"}`)
	metadata := map[string]string{
		"content-type": "application/json",
		"version":      "1.0",
		"author":       "Test Author",
	}

	err = hub.PutToObjectStore("documents", "doc1.json", documentData, metadata)
	if err != nil {
		t.Fatalf("Failed to put object: %v", err)
	}

	// Test GET operation
	retrievedData, err := hub.GetFromObjectStore("documents", "doc1.json")
	if err != nil {
		t.Fatalf("Failed to get object: %v", err)
	}

	if string(retrievedData) != string(documentData) {
		t.Errorf("Retrieved data doesn't match original")
	}

	// Test PUT with different metadata
	updatedData := []byte(`{"title":"Updated Document","content":"This is an updated test document","author":"Test Author"}`)
	updatedMetadata := map[string]string{
		"content-type":  "application/json",
		"version":       "2.0",
		"author":        "Test Author",
		"last-modified": time.Now().Format(time.RFC3339),
	}

	err = hub.PutToObjectStore("documents", "doc1.json", updatedData, updatedMetadata)
	if err != nil {
		t.Fatalf("Failed to update object: %v", err)
	}

	// Verify update
	retrievedData, err = hub.GetFromObjectStore("documents", "doc1.json")
	if err != nil {
		t.Fatalf("Failed to get updated object: %v", err)
	}

	if string(retrievedData) != string(updatedData) {
		t.Errorf("Updated data doesn't match")
	}

	// Test DELETE operation
	err = hub.DeleteFromObjectStore("documents", "doc1.json")
	if err != nil {
		t.Fatalf("Failed to delete object: %v", err)
	}

	// Verify deletion
	_, err = hub.GetFromObjectStore("documents", "doc1.json")
	if err == nil {
		t.Error("Expected error when getting deleted object")
	}

	t.Logf("Successfully completed CRUD operations")
}

func TestObjectStoreConcurrency(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "obj_test_concurrency_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
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
	defer hub.Shutdown()

	// Create object store
	config := ObjectStoreConfig{
		Bucket:      "concurrent-objects",
		Description: "Concurrent access test store",
		TTL:         time.Hour,
		MaxBytes:    NewSizeFromMegabytes(100),
		Replicas:    1,
	}
	err = hub.CreateObjectStore(config)
	if err != nil {
		t.Fatalf("Failed to create object store: %v", err)
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
				key := fmt.Sprintf("object-%d-%d.txt", goroutineID, j)
				data := []byte(fmt.Sprintf("Content from goroutine %d, operation %d", goroutineID, j))
				metadata := map[string]string{
					"goroutine": fmt.Sprintf("%d", goroutineID),
					"operation": fmt.Sprintf("%d", j),
				}

				// PUT operation
				err := hub.PutToObjectStore("concurrent-objects", key, data, metadata)
				if err != nil {
					t.Errorf("Goroutine %d failed to put: %v", goroutineID, err)
					continue
				}

				// GET operation
				retrievedData, err := hub.GetFromObjectStore("concurrent-objects", key)
				if err != nil {
					t.Errorf("Goroutine %d failed to get: %v", goroutineID, err)
					continue
				}

				if string(retrievedData) != string(data) {
					t.Errorf("Goroutine %d: data mismatch", goroutineID)
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

func TestObjectStoreClusterOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "obj_test_cluster_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
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
	defer hub.Shutdown()

	// Create multiple object stores for cluster simulation
	stores := []string{"node1-objects", "node2-objects", "node3-objects"}

	for _, storeName := range stores {
		config := ObjectStoreConfig{
			Bucket:      storeName,
			Description: fmt.Sprintf("Object store for %s", storeName),
			TTL:         time.Hour,
			MaxBytes:    NewSizeFromMegabytes(50),
			Replicas:    1,
		}

		err = hub.CreateObjectStore(config)
		if err != nil {
			t.Fatalf("Failed to create object store %s: %v", storeName, err)
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
			key := fmt.Sprintf("node%d-file.txt", nodeID)
			data := []byte(fmt.Sprintf("Data from node %d in cluster", nodeID))
			metadata := map[string]string{
				"node":      fmt.Sprintf("%d", nodeID),
				"cluster":   "test-cluster",
				"timestamp": time.Now().Format(time.RFC3339),
			}

			// PUT operation
			err := hub.PutToObjectStore(store, key, data, metadata)
			if err != nil {
				results <- fmt.Sprintf("PUT failed for %s: %v", store, err)
				return
			}

			// GET operation
			retrievedData, err := hub.GetFromObjectStore(store, key)
			if err != nil {
				results <- fmt.Sprintf("GET failed for %s: %v", store, err)
				return
			}

			if string(retrievedData) != string(data) {
				results <- fmt.Sprintf("Data mismatch for %s", store)
				return
			}

			// UPDATE operation with new metadata
			updatedData := []byte(fmt.Sprintf("Updated data from node %d", nodeID))
			updatedMetadata := map[string]string{
				"node":       fmt.Sprintf("%d", nodeID),
				"cluster":    "test-cluster",
				"updated":    "true",
				"updated_at": time.Now().Format(time.RFC3339),
			}

			err = hub.PutToObjectStore(store, key, updatedData, updatedMetadata)
			if err != nil {
				results <- fmt.Sprintf("UPDATE failed for %s: %v", store, err)
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

func TestObjectStoreMetadata(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "obj_test_metadata_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
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
	defer hub.Shutdown()

	// Create object store with metadata
	config := ObjectStoreConfig{
		Bucket:      "metadata-test",
		Description: "Metadata test store",
		TTL:         time.Hour,
		MaxBytes:    NewSizeFromMegabytes(10),
		Replicas:    1,
		Metadata: map[string]string{
			"store-type":  "metadata-test",
			"environment": "test",
			"created-by":  "test-suite",
		},
	}
	err = hub.CreateObjectStore(config)
	if err != nil {
		t.Fatalf("Failed to create object store: %v", err)
	}

	// Test storing object with rich metadata
	fileData := []byte("This is test file content for metadata testing")
	fileMetadata := map[string]string{
		"filename":     "test.txt",
		"content-type": "text/plain",
		"size":         fmt.Sprintf("%d", len(fileData)),
		"checksum":     "abc123",
		"tags":         "test,metadata,object-store",
		"version":      "1.0.0",
		"author":       "Test Suite",
		"created":      time.Now().Format(time.RFC3339),
	}

	err = hub.PutToObjectStore("metadata-test", "test.txt", fileData, fileMetadata)
	if err != nil {
		t.Fatalf("Failed to put object with metadata: %v", err)
	}

	// Verify data can be retrieved
	retrievedData, err := hub.GetFromObjectStore("metadata-test", "test.txt")
	if err != nil {
		t.Fatalf("Failed to get object: %v", err)
	}

	if string(retrievedData) != string(fileData) {
		t.Errorf("Retrieved data doesn't match original")
	}

	// Test updating with different metadata
	updatedData := []byte("Updated content for metadata testing")
	updatedMetadata := map[string]string{
		"filename":     "test.txt",
		"content-type": "text/plain",
		"size":         fmt.Sprintf("%d", len(updatedData)),
		"checksum":     "def456",
		"tags":         "test,metadata,updated",
		"version":      "1.1.0",
		"author":       "Test Suite",
		"updated":      time.Now().Format(time.RFC3339),
	}

	err = hub.PutToObjectStore("metadata-test", "test.txt", updatedData, updatedMetadata)
	if err != nil {
		t.Fatalf("Failed to update object with metadata: %v", err)
	}

	// Verify updated data
	retrievedData, err = hub.GetFromObjectStore("metadata-test", "test.txt")
	if err != nil {
		t.Fatalf("Failed to get updated object: %v", err)
	}

	if string(retrievedData) != string(updatedData) {
		t.Errorf("Updated data doesn't match")
	}

	t.Logf("Successfully tested metadata functionality")
}

func TestObjectStoreErrorHandling(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "obj_test_error_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
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
	defer hub.Shutdown()

	// Test operations on non-existent bucket
	_, err = hub.GetFromObjectStore("nonexistent-bucket", "test-object")
	if err == nil {
		t.Error("Expected error when accessing non-existent bucket")
	} else {
		t.Logf("Got expected error for non-existent bucket: %v", err)
	}

	err = hub.PutToObjectStore("nonexistent-bucket", "test-object", []byte("test"), nil)
	if err == nil {
		t.Error("Expected error when putting to non-existent bucket")
	} else {
		t.Logf("Got expected error for put to non-existent bucket: %v", err)
	}

	// Test operations on non-existent object
	// First create a valid bucket
	config := ObjectStoreConfig{
		Bucket:      "error-test",
		Description: "Error handling test store",
		TTL:         time.Hour,
		MaxBytes:    NewSizeFromMegabytes(10),
		Replicas:    1,
	}
	err = hub.CreateObjectStore(config)
	if err != nil {
		t.Fatalf("Failed to create object store: %v", err)
	}

	_, err = hub.GetFromObjectStore("error-test", "nonexistent-object")
	if err == nil {
		t.Error("Expected error when getting non-existent object")
	} else {
		t.Logf("Got expected error for non-existent object: %v", err)
	}

	// Test delete non-existent object
	err = hub.DeleteFromObjectStore("error-test", "nonexistent-object")
	if err == nil {
		t.Error("Expected error when deleting non-existent object")
	} else {
		t.Logf("Got expected error for delete non-existent object: %v", err)
	}

	// Test storing and retrieving valid data
	validData := []byte("Valid test data")
	err = hub.PutToObjectStore("error-test", "valid-object", validData, nil)
	if err != nil {
		t.Fatalf("Failed to put valid data: %v", err)
	}

	retrievedData, err := hub.GetFromObjectStore("error-test", "valid-object")
	if err != nil {
		t.Fatalf("Failed to get valid data: %v", err)
	}

	if string(retrievedData) != string(validData) {
		t.Errorf("Valid data retrieval failed")
	}

	t.Logf("Successfully tested error handling scenarios")
}
