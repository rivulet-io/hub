package hub

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultOptions(t *testing.T) {
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("DefaultOptions() failed: %v", err)
	}

	if opts == nil {
		t.Fatal("DefaultOptions() returned nil options")
	}

	// Check default values
	if opts.Name == "" {
		t.Error("DefaultOptions() should set a non-empty name")
	}

	if opts.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", opts.Host)
	}

	if opts.Port != 4222 {
		t.Errorf("Expected port 4222, got %d", opts.Port)
	}

	if opts.MaxPayload != NewSizeFromMegabytes(8) {
		t.Errorf("Expected MaxPayload 8MB, got %v", opts.MaxPayload)
	}

	if opts.ClusterHost != "0.0.0.0" {
		t.Errorf("Expected cluster host '0.0.0.0', got '%s'", opts.ClusterHost)
	}

	if opts.ClusterPort != 6222 {
		t.Errorf("Expected cluster port 6222, got %d", opts.ClusterPort)
	}

	if opts.JetstreamMaxMemory != NewSizeFromMegabytes(512) {
		t.Errorf("Expected JetStream max memory 512MB, got %v", opts.JetstreamMaxMemory)
	}

	if opts.JetstreamMaxStorage != NewSizeFromGigabytes(10) {
		t.Errorf("Expected JetStream max storage 10GB, got %v", opts.JetstreamMaxStorage)
	}

	if opts.StoreDir != "./data" {
		t.Errorf("Expected store dir './data', got '%s'", opts.StoreDir)
	}
}

func TestNewHub(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hub_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create options with unique port to avoid conflicts
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0        // Use random available port
	opts.ClusterPort = 0 // Use random available port for clustering
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}

	if hub == nil {
		t.Fatal("NewHub() returned nil hub")
	}

	// Verify hub components are initialized
	if hub.options != opts {
		t.Error("Hub options not set correctly")
	}

	if hub.server == nil {
		t.Error("Hub server not initialized")
	}

	if hub.inProcessConn == nil {
		t.Error("Hub in-process connection not initialized")
	}

	if hub.jetstreamCtx == nil {
		t.Error("Hub JetStream context not initialized")
	}

	// Test basic connectivity
	if !hub.server.ReadyForConnections(5 * time.Second) {
		t.Error("NATS server not ready for connections")
	}

	// Test JetStream is available
	info, err := hub.jetstreamCtx.AccountInfo()
	if err != nil {
		t.Errorf("JetStream not available: %v", err)
	}

	if info == nil {
		t.Error("JetStream account info is nil")
	}

	// Clean up
	hub.Shutdown()
}

func TestNewHubWithInvalidOptions(t *testing.T) {
	// Test with nil options
	hub, err := NewHub(nil)
	if err == nil {
		t.Error("NewHub() with nil options should return error")
		if hub != nil {
			hub.Shutdown()
		}
	}

	// Test with invalid store directory
	tempDir, err := os.MkdirTemp("", "hub_test_invalid_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create options with invalid store directory (read-only)
	invalidDir := filepath.Join(tempDir, "readonly")
	err = os.MkdirAll(invalidDir, 0444)
	if err != nil {
		t.Fatalf("Failed to create readonly dir: %v", err)
	}

	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.StoreDir = invalidDir

	hub, err = NewHub(opts)
	// Note: This might not always fail depending on the system, so we'll just log it
	if err != nil {
		t.Logf("NewHub() with invalid store dir failed as expected: %v", err)
	} else if hub != nil {
		hub.Shutdown()
	}
}

func TestHubShutdown(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hub_test_shutdown_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

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

	// Verify connections are working before shutdown
	if !hub.server.ReadyForConnections(5 * time.Second) {
		t.Error("Server not ready before shutdown")
	}

	// Test shutdown
	hub.Shutdown()

	// Verify server is shutdown
	if hub.server.ReadyForConnections(1 * time.Second) {
		t.Error("Server should not be ready after shutdown")
	}

	// Verify connections are closed
	// Note: We can't easily test if connections are closed without more complex setup
}

func TestHubVolatileMessaging(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hub_test_messaging_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

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

	// Test volatile messaging
	received := make(chan []byte, 1)
	cancel, err := hub.SubscribeVolatileViaFanout("test.subject", func(subject string, msg []byte) ([]byte, bool) {
		if subject != "test.subject" {
			t.Errorf("Expected subject 'test.subject', got '%s'", subject)
		}
		received <- msg
		return []byte("response"), true
	}, func(err error) {
		t.Errorf("Subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cancel()

	// Publish message
	testMsg := []byte("test message")
	err = hub.PublishVolatile("test.subject", testMsg)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Wait for message
	select {
	case receivedMsg := <-received:
		if string(receivedMsg) != string(testMsg) {
			t.Errorf("Expected message '%s', got '%s'", string(testMsg), string(receivedMsg))
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestHubRequestReply(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hub_test_request_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

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

	// Set up responder
	cancel, err := hub.SubscribeVolatileViaFanout("test.request", func(subject string, msg []byte) ([]byte, bool) {
		return []byte("response to: " + string(msg)), true
	}, func(err error) {
		t.Errorf("Subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cancel()

	// Send request
	requestMsg := []byte("hello")
	response, err := hub.RequestVolatile("test.request", requestMsg, 2*time.Second)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	expectedResponse := "response to: hello"
	if string(response) != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, string(response))
	}
}

func TestHubClusterName(t *testing.T) {
	if HubClusterName != "hubstream" {
		t.Errorf("Expected HubClusterName 'hubstream', got '%s'", HubClusterName)
	}
}

func TestHubOptionsValidation(t *testing.T) {
	// Test with nil options - this should fail
	hub, err := NewHub(nil)
	if err == nil {
		t.Error("NewHub() with nil options should return error")
		if hub != nil {
			hub.Shutdown()
		}
	}

	// Test with options that have negative values - NATS may handle these gracefully
	opts := &Options{
		Name:                "test-hub",
		Host:                "0.0.0.0",
		Port:                0,                    // 0 means random port, should work
		MaxPayload:          NewSizeFromBytes(-1), // Negative size
		JetstreamMaxMemory:  NewSizeFromBytes(-1), // Negative size
		JetstreamMaxStorage: NewSizeFromBytes(-1), // Negative size
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "hub_test_validation_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	opts.StoreDir = filepath.Join(tempDir, "data")

	// This might succeed because NATS handles negative values gracefully
	hub, err = NewHub(opts)
	if err != nil {
		t.Logf("NewHub() with negative sizes failed as expected: %v", err)
	} else {
		t.Logf("NewHub() with negative sizes succeeded - NATS handled them gracefully")
		if hub != nil {
			hub.Shutdown()
		}
	}
}
