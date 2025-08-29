package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/snowmerak/hub"
)

// ClusterNode represents a node in the cluster
type ClusterNode struct {
	ID          string
	Hub         *hub.Hub
	StoreDir    string
	Port        int
	ClusterPort int
	Options     *hub.Options
}

func clusterTestFunc() {
	fmt.Println("=== Hub Cluster Integration Tests ===")

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hub_cluster_test_*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Using temp directory: %s\n", tempDir)

	// Run all cluster tests
	tests := []struct {
		name string
		fn   func(string) error
	}{
		{"Cluster Formation", testClusterFormation},
		{"Cluster Basic Communication", testClusterBasicCommunication},
		{"Cluster JetStream Operations", testClusterJetStreamOperations},
		{"Cluster Key-Value Store", testClusterKeyValueStore},
		{"Cluster Object Store", testClusterObjectStore},
		{"Cluster Load Balancing", testClusterLoadBalancing},
		{"Cluster Error Handling", testClusterErrorHandling},
		{"Cluster Node Failure Recovery", testClusterNodeFailureRecovery},
		{"Cluster Concurrent Operations", testClusterConcurrentOperations},
		{"Cluster Performance", testClusterPerformance},
		{"Cluster Data Consistency", testClusterDataConsistency},
		{"Cluster Network Resilience", testClusterNetworkResilience},
		{"Cluster Stream Replication", testClusterStreamReplication},
		{"Cluster Consumer Management", testClusterConsumerManagement},
		{"Cluster Configuration Validation", testClusterConfigurationValidation},
	}

	passed := 0
	failed := 0

	for _, test := range tests {
		fmt.Printf("\n--- Running %s ---\n", test.name)
		if err := test.fn(tempDir); err != nil {
			fmt.Printf("❌ FAILED: %v\n", err)
			failed++
		} else {
			fmt.Printf("✅ PASSED\n")
			passed++
		}
	}

	fmt.Printf("\n=== Cluster Test Results ===\n")
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Total: %d\n", passed+failed)

	if failed > 0 {
		os.Exit(1)
	}
}

// createClusterNode creates a new cluster node with specified configuration
func createClusterNode(nodeID string, tempDir string, port, clusterPort int, routes []*url.URL) (*ClusterNode, error) {
	storeDir := filepath.Join(tempDir, fmt.Sprintf("node_%s", nodeID))

	opts := &hub.Options{
		Name:       nodeID,
		Host:       "127.0.0.1",
		Port:       port,
		MaxPayload: hub.NewSizeFromMegabytes(8),

		Routes: routes,

		ClusterHost:         "127.0.0.1",
		ClusterPort:         clusterPort,
		ClusterConnPoolSize: 4,
		ClusterPingInterval: 30 * time.Second,

		// Disable gateway and leafnode for cluster testing
		GatewayHost: "",
		GatewayPort: 0,

		LeafNodeHost: "",
		LeafNodePort: 0,

		JetstreamMaxMemory:    hub.NewSizeFromMegabytes(256),
		JetstreamMaxStorage:   hub.NewSizeFromMegabytes(512),
		StreamMaxBufferedMsgs: 1000,
		StreamMaxBufferedSize: 64 * 1024 * 1024,
		StoreDir:              storeDir,
		SyncInterval:          2 * time.Second,
		SyncAlways:            false,

		LogSizeLimit: 10 * 1024 * 1024,
		LogMaxFiles:  3,
		Syslog:       false,
		RemoteSyslog: "",

		ClientAuthenticationMethod: nil,
		RouterAuthenticationMethod: nil,
	}

	h, err := hub.NewHub(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create hub for node %s: %w", nodeID, err)
	}

	return &ClusterNode{
		ID:          nodeID,
		Hub:         h,
		StoreDir:    storeDir,
		Port:        port,
		ClusterPort: clusterPort,
		Options:     opts,
	}, nil
}

// createCluster creates a cluster of N nodes
func createCluster(nodeCount int, tempDir string) ([]*ClusterNode, error) {
	nodes := make([]*ClusterNode, nodeCount)

	basePort := 14222
	baseClusterPort := 16222

	// First pass: create all nodes with routes to first node
	for i := 0; i < nodeCount; i++ {
		nodeID := fmt.Sprintf("cluster-node-%d", i+1)
		port := basePort + i
		clusterPort := baseClusterPort + i

		var routes []*url.URL
		if i > 0 {
			// Connect to the first node
			routes = []*url.URL{{Scheme: "nats", Host: fmt.Sprintf("127.0.0.1:%d", baseClusterPort)}}
		}

		node, err := createClusterNode(nodeID, tempDir, port, clusterPort, routes)
		if err != nil {
			// Cleanup already created nodes
			for j := 0; j < i; j++ {
				if nodes[j] != nil {
					nodes[j].Hub.Shutdown()
				}
			}
			return nil, err
		}

		nodes[i] = node
		fmt.Printf("✓ Created cluster node %s (port: %d, cluster port: %d)\n", nodeID, port, clusterPort)
	}

	// Wait for cluster formation
	fmt.Println("Waiting for cluster formation...")
	time.Sleep(8 * time.Second)

	return nodes, nil
}

// shutdownCluster shuts down all nodes in the cluster
func shutdownCluster(nodes []*ClusterNode) {
	for _, node := range nodes {
		if node != nil && node.Hub != nil {
			fmt.Printf("Shutting down node %s...\n", node.ID)
			node.Hub.Shutdown()
		}
	}
}

func testClusterFormation(tempDir string) error {
	fmt.Println("Testing cluster formation...")

	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Printf("✓ Successfully formed %d-node cluster\n", nodeCount)

	// Verify all nodes are operational
	for _, node := range nodes {
		// Test basic functionality
		testMsg := []byte(fmt.Sprintf("test from %s", node.ID))
		err := node.Hub.PublishVolatile("cluster.formation.test", testMsg)
		if err != nil {
			return fmt.Errorf("node %s is not operational: %w", node.ID, err)
		}
	}

	fmt.Printf("✓ All %d nodes are operational\n", nodeCount)
	return nil
}

func testClusterBasicCommunication(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing basic cluster communication...")

	// Test cross-node communication
	received := make(chan struct{}, 10)
	var mu sync.Mutex
	receivedCount := 0

	// Subscribe on all nodes
	cancels := make([]func(), 0)
	for _, node := range nodes {
		cancel, err := node.Hub.SubscribeVolatileViaFanout("cluster.broadcast", func(subject string, msg []byte) ([]byte, bool) {
			mu.Lock()
			receivedCount++
			mu.Unlock()
			received <- struct{}{}
			fmt.Printf("Node %s received: %s\n", node.ID, string(msg))
			return nil, false
		}, func(err error) {
			log.Printf("Subscription error on %s: %v", node.ID, err)
		})

		if err != nil {
			for _, c := range cancels {
				c()
			}
			return fmt.Errorf("failed to subscribe on %s: %w", node.ID, err)
		}
		cancels = append(cancels, cancel)
	}
	defer func() {
		for _, cancel := range cancels {
			cancel()
		}
	}()

	// Wait for subscriptions to be established
	time.Sleep(2 * time.Second)

	// Publish from first node
	testMsg := []byte("Hello cluster!")
	err = nodes[0].Hub.PublishVolatile("cluster.broadcast", testMsg)
	if err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	// Wait for messages to be received
	timeout := time.After(5 * time.Second)
	expectedReceives := len(nodes) // Each node should receive the message

	for i := 0; i < expectedReceives; i++ {
		select {
		case <-received:
			// Message received
		case <-timeout:
			return fmt.Errorf("timeout waiting for message %d, received %d", i+1, receivedCount)
		}
	}

	mu.Lock()
	if receivedCount != expectedReceives {
		return fmt.Errorf("expected %d receives, got %d", expectedReceives, receivedCount)
	}
	mu.Unlock()

	fmt.Printf("✓ Successfully tested cluster communication (%d messages received)\n", receivedCount)
	return nil
}

func testClusterJetStreamOperations(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing JetStream operations across cluster...")

	// Create stream on first node
	config := &hub.PersistentConfig{
		Description:  "Cluster JetStream test stream",
		Subjects:     []string{"cluster.jetstream.>"},
		Retention:    0, // LimitsPolicy
		MaxConsumers: 10,
		MaxMsgs:      1000,
		MaxBytes:     hub.NewSizeFromMegabytes(10).Bytes(),
		MaxAge:       24 * time.Hour,
		Replicas:     3, // Replicate across cluster
	}

	err = nodes[0].Hub.CreateOrUpdatePersistent(config)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// Publish messages from different nodes
	messages := []string{"Message 1", "Message 2", "Message 3"}
	for i, msg := range messages {
		publisherNode := nodes[i%len(nodes)]
		subject := fmt.Sprintf("cluster.jetstream.msg%d", i+1)
		message := []byte(fmt.Sprintf("[%s] %s", publisherNode.ID, msg))

		err := publisherNode.Hub.PublishPersistent(subject, message)
		if err != nil {
			return fmt.Errorf("failed to publish from %s: %w", publisherNode.ID, err)
		}
		fmt.Printf("✓ Published from %s: %s\n", publisherNode.ID, string(message))
	}

	// Subscribe and consume from a different node
	subscriberNode := nodes[len(nodes)-1] // Last node
	received := make(chan []byte, len(messages))
	var receivedCount int
	var mu sync.Mutex

	cancel, err := subscriberNode.Hub.SubscribePersistentViaDurable("cluster-consumer", "cluster.jetstream.>", func(subject string, msg []byte) ([]byte, bool, bool) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		received <- msg
		return nil, false, true
	}, func(err error) {
		log.Printf("Subscription error: %v", err)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	defer cancel()

	// Wait for all messages
	timeout := time.After(10 * time.Second)
	for i := 0; i < len(messages); i++ {
		select {
		case msg := <-received:
			fmt.Printf("✓ Received: %s\n", string(msg))
		case <-timeout:
			return fmt.Errorf("timeout waiting for message %d", i+1)
		}
	}

	mu.Lock()
	if receivedCount != len(messages) {
		return fmt.Errorf("expected %d messages, got %d", len(messages), receivedCount)
	}
	mu.Unlock()

	fmt.Printf("✓ JetStream cluster operations successful (%d messages)\n", len(messages))
	return nil
}

func testClusterKeyValueStore(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing Key-Value store across cluster...")

	// Create KV store on first node
	kvConfig := hub.KeyValueStoreConfig{
		Bucket:       "cluster_kv_test",
		Description:  "Cluster KV store test",
		MaxValueSize: hub.NewSizeFromKilobytes(64),
		TTL:          time.Hour,
		MaxBytes:     hub.NewSizeFromMegabytes(10),
		Replicas:     3,
	}

	err = nodes[0].Hub.CreateOrUpdateKeyValueStore(kvConfig)
	if err != nil {
		return fmt.Errorf("failed to create KV store: %w", err)
	}

	// Put values from different nodes
	testData := map[string]string{
		"cluster_key_1": "cluster_value_1",
		"cluster_key_2": "cluster_value_2",
		"cluster_key_3": "cluster_value_3",
	}

	for i, key := range []string{"cluster_key_1", "cluster_key_2", "cluster_key_3"} {
		putter := nodes[i%len(nodes)]
		value := testData[key]
		_, err := putter.Hub.PutToKeyValueStore("cluster_kv_test", key, []byte(value))
		if err != nil {
			return fmt.Errorf("failed to put key %s from %s: %w", key, putter.ID, err)
		}
		fmt.Printf("✓ Put key %s from %s: %s\n", key, putter.ID, value)
	}

	// Wait for replication
	time.Sleep(2 * time.Second)

	// Get values from different nodes
	for i, key := range []string{"cluster_key_1", "cluster_key_2", "cluster_key_3"} {
		getter := nodes[(i+1)%len(nodes)] // Different node than putter
		expectedValue := testData[key]

		value, _, err := getter.Hub.GetFromKeyValueStore("cluster_kv_test", key)
		if err != nil {
			return fmt.Errorf("failed to get key %s from %s: %w", key, getter.ID, err)
		}

		if string(value) != expectedValue {
			return fmt.Errorf("key %s on %s: expected %s, got %s", key, getter.ID, expectedValue, string(value))
		}

		fmt.Printf("✓ Got key %s from %s: %s\n", key, getter.ID, string(value))
	}

	fmt.Printf("✓ Cluster KV store operations successful\n")
	return nil
}

func testClusterObjectStore(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing Object store across cluster...")

	// Create object store on first node
	objConfig := hub.ObjectStoreConfig{
		Bucket:      "cluster_objects",
		Description: "Cluster object store test",
		TTL:         time.Hour,
		MaxBytes:    hub.NewSizeFromMegabytes(100),
		Replicas:    3,
		Metadata:    map[string]string{"environment": "cluster-test"},
	}

	err = nodes[0].Hub.CreateObjectStore(objConfig)
	if err != nil {
		return fmt.Errorf("failed to create object store: %w", err)
	}

	// Put objects from different nodes
	testObjects := map[string][]byte{
		"object1.txt": []byte("Cluster object data 1"),
		"object2.txt": []byte("Cluster object data 2"),
		"object3.txt": []byte("Cluster object data 3"),
	}

	objectNames := []string{"object1.txt", "object2.txt", "object3.txt"}
	for i, objName := range objectNames {
		putter := nodes[i%len(nodes)]
		data := testObjects[objName]

		err := putter.Hub.PutToObjectStore("cluster_objects", objName, data, map[string]string{"content-type": "text/plain"})
		if err != nil {
			return fmt.Errorf("failed to put object %s from %s: %w", objName, putter.ID, err)
		}
		fmt.Printf("✓ Put object %s from %s (%d bytes)\n", objName, putter.ID, len(data))
	}

	// Wait for replication
	time.Sleep(2 * time.Second)

	// Get objects from different nodes
	for i, objName := range objectNames {
		getter := nodes[(i+1)%len(nodes)] // Different node than putter
		expectedData := testObjects[objName]

		retrievedData, err := getter.Hub.GetFromObjectStore("cluster_objects", objName)
		if err != nil {
			return fmt.Errorf("failed to get object %s from %s: %w", objName, getter.ID, err)
		}

		if string(retrievedData) != string(expectedData) {
			return fmt.Errorf("object %s on %s: data mismatch", objName, getter.ID)
		}

		fmt.Printf("✓ Got object %s from %s (%d bytes)\n", objName, getter.ID, len(retrievedData))
	}

	fmt.Printf("✓ Cluster object store operations successful\n")
	return nil
}

func testClusterLoadBalancing(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing load balancing across cluster...")

	// Create work queue stream
	config := &hub.PersistentConfig{
		Description: "Cluster load balancing test stream",
		Subjects:    []string{"cluster.work.>"},
		Retention:   0,
		MaxMsgs:     100,
		Replicas:    3,
	}

	err = nodes[0].Hub.CreateOrUpdatePersistent(config)
	if err != nil {
		return fmt.Errorf("failed to create work queue stream: %w", err)
	}

	// Create workers on each node
	workerResults := make(map[string]int)
	var mu sync.Mutex
	cancelFuncs := make([]func(), 0)

	for _, node := range nodes {
		workerID := fmt.Sprintf("worker-%s", node.ID)

		cancel, err := node.Hub.SubscribePersistentViaDurable(workerID, "cluster.work.tasks", func(subject string, msg []byte) ([]byte, bool, bool) {
			mu.Lock()
			workerResults[workerID]++
			mu.Unlock()
			fmt.Printf("Worker %s processed: %s\n", workerID, string(msg))
			return nil, false, true
		}, func(err error) {
			log.Printf("Worker %s error: %v", workerID, err)
		})

		if err != nil {
			for _, c := range cancelFuncs {
				c()
			}
			return fmt.Errorf("failed to create worker %s: %w", workerID, err)
		}
		cancelFuncs = append(cancelFuncs, cancel)
	}
	defer func() {
		for _, cancel := range cancelFuncs {
			cancel()
		}
	}()

	// Wait for workers to be ready
	time.Sleep(3 * time.Second)

	// Publish work items
	workItems := 9
	for i := 0; i < workItems; i++ {
		publisher := nodes[i%len(nodes)]
		msg := []byte(fmt.Sprintf("work-item-%d", i+1))
		err := publisher.Hub.PublishPersistent("cluster.work.tasks", msg)
		if err != nil {
			return fmt.Errorf("failed to publish work item %d: %w", i+1, err)
		}
	}

	// Wait for processing
	time.Sleep(5 * time.Second)

	// Check load distribution
	mu.Lock()
	totalProcessed := 0
	for workerID, count := range workerResults {
		fmt.Printf("Worker %s processed %d items\n", workerID, count)
		totalProcessed += count
	}
	mu.Unlock()

	if totalProcessed != workItems {
		return fmt.Errorf("expected %d items processed, got %d", workItems, totalProcessed)
	}

	fmt.Printf("✓ Load balancing test successful (%d items distributed)\n", totalProcessed)
	return nil
}

func testClusterErrorHandling(tempDir string) error {
	nodeCount := 2
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing cluster error handling...")

	// Test 1: Publishing to non-existent stream
	err = nodes[0].Hub.PublishPersistent("nonexistent.stream", []byte("test"))
	if err == nil {
		return fmt.Errorf("expected error when publishing to non-existent stream")
	}
	fmt.Printf("✓ Correctly handled non-existent stream: %v\n", err)

	// Test 2: Invalid KV operations
	_, _, err = nodes[1].Hub.GetFromKeyValueStore("nonexistent_bucket", "nonexistent_key")
	if err == nil {
		return fmt.Errorf("expected error when getting non-existent KV key")
	}
	fmt.Printf("✓ Correctly handled non-existent KV key: %v\n", err)

	// Test 3: Invalid object operations
	_, err = nodes[0].Hub.GetFromObjectStore("nonexistent_bucket", "nonexistent_object")
	if err == nil {
		return fmt.Errorf("expected error when getting non-existent object")
	}
	fmt.Printf("✓ Correctly handled non-existent object: %v\n", err)

	fmt.Println("✓ Cluster error handling test passed")
	return nil
}

func testClusterNodeFailureRecovery(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing cluster node failure recovery...")

	// Create replicated stream
	config := &hub.PersistentConfig{
		Description: "Node failure recovery test stream",
		Subjects:    []string{"cluster.recovery.>"},
		Retention:   0,
		MaxMsgs:     100,
		Replicas:    3,
	}

	err = nodes[0].Hub.CreateOrUpdatePersistent(config)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// Publish some messages
	for i := 0; i < 5; i++ {
		msg := []byte(fmt.Sprintf("Recovery message %d", i+1))
		err := nodes[0].Hub.PublishPersistent("cluster.recovery.msg", msg)
		if err != nil {
			return fmt.Errorf("failed to publish message %d: %w", i+1, err)
		}
	}

	// Simulate node failure by shutting down one node
	fmt.Printf("Simulating failure of node %s\n", nodes[1].ID)
	nodes[1].Hub.Shutdown()

	// Wait for cluster to adjust
	time.Sleep(3 * time.Second)

	// Verify remaining nodes can still operate
	err = nodes[0].Hub.PublishPersistent("cluster.recovery.post-failure", []byte("Post-failure message"))
	if err != nil {
		return fmt.Errorf("failed to publish after node failure: %w", err)
	}

	// Subscribe and verify message delivery still works
	received := make(chan []byte, 10)
	cancel, err := nodes[2].Hub.SubscribePersistentViaDurable("recovery-consumer", "cluster.recovery.>", func(subject string, msg []byte) ([]byte, bool, bool) {
		received <- msg
		return nil, false, true
	}, func(err error) {
		log.Printf("Subscription error: %v", err)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe after failure: %w", err)
	}
	defer cancel()

	// Wait for at least the post-failure message
	timeout := time.After(5 * time.Second)
	receivedCount := 0
	for {
		select {
		case msg := <-received:
			receivedCount++
			fmt.Printf("✓ Received after failure: %s\n", string(msg))
			if string(msg) == "Post-failure message" {
				fmt.Printf("✓ Node failure recovery test successful (received %d messages)\n", receivedCount)
				return nil
			}
		case <-timeout:
			return fmt.Errorf("timeout waiting for post-failure message, received %d messages", receivedCount)
		}
	}
}

func testClusterConcurrentOperations(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing concurrent operations across cluster...")

	// Create KV store for concurrent testing
	kvConfig := hub.KeyValueStoreConfig{
		Bucket:   "cluster_concurrent_test",
		Replicas: 3,
	}
	err = nodes[0].Hub.CreateOrUpdateKeyValueStore(kvConfig)
	if err != nil {
		return fmt.Errorf("failed to create KV store: %w", err)
	}

	// Wait for replication
	time.Sleep(2 * time.Second)

	// Test concurrent KV operations across nodes
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)
	successCount := 0

	operationsPerNode := 5
	for nodeIdx, node := range nodes {
		for i := 0; i < operationsPerNode; i++ {
			wg.Add(1)
			go func(node *ClusterNode, nodeIdx, opIdx int) {
				defer wg.Done()

				key := fmt.Sprintf("concurrent_key_%d_%d", nodeIdx, opIdx)
				value := fmt.Sprintf("concurrent_value_%d_%d", nodeIdx, opIdx)

				// Put operation
				_, err := node.Hub.PutToKeyValueStore("cluster_concurrent_test", key, []byte(value))
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("put error for %s on %s: %w", key, node.ID, err))
					mu.Unlock()
					return
				}

				// Get operation
				retrieved, _, err := node.Hub.GetFromKeyValueStore("cluster_concurrent_test", key)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("get error for %s on %s: %w", key, node.ID, err))
					mu.Unlock()
					return
				}

				if string(retrieved) != value {
					mu.Lock()
					errors = append(errors, fmt.Errorf("value mismatch for %s on %s: expected %s, got %s", key, node.ID, value, string(retrieved)))
					mu.Unlock()
					return
				}

				mu.Lock()
				successCount++
				mu.Unlock()
			}(node, nodeIdx, i)
		}
	}

	wg.Wait()

	mu.Lock()
	if len(errors) > 0 {
		return fmt.Errorf("concurrent operations failed: %v", errors[0])
	}
	mu.Unlock()

	expectedOperations := len(nodes) * operationsPerNode
	if successCount != expectedOperations {
		return fmt.Errorf("expected %d successful operations, got %d", expectedOperations, successCount)
	}

	fmt.Printf("✓ Successfully completed %d concurrent operations across cluster\n", successCount)
	return nil
}

func testClusterPerformance(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing cluster performance...")

	// Create KV store for performance testing
	kvConfig := hub.KeyValueStoreConfig{
		Bucket:   "cluster_perf_test",
		Replicas: 3,
	}
	err = nodes[0].Hub.CreateOrUpdateKeyValueStore(kvConfig)
	if err != nil {
		return fmt.Errorf("failed to create KV store: %w", err)
	}

	// Test KV put performance across nodes
	operations := 50
	start := time.Now()

	for i := 0; i < operations; i++ {
		node := nodes[i%len(nodes)]
		key := fmt.Sprintf("perf_key_%d", i)
		value := fmt.Sprintf("perf_value_%d", i)
		_, err := node.Hub.PutToKeyValueStore("cluster_perf_test", key, []byte(value))
		if err != nil {
			return fmt.Errorf("KV put failed at operation %d on %s: %w", i, node.ID, err)
		}
	}
	kvPutDuration := time.Since(start)

	// Test KV get performance across nodes
	start = time.Now()
	for i := 0; i < operations; i++ {
		node := nodes[(i+1)%len(nodes)] // Different node pattern
		key := fmt.Sprintf("perf_key_%d", i)
		_, _, err := node.Hub.GetFromKeyValueStore("cluster_perf_test", key)
		if err != nil {
			return fmt.Errorf("KV get failed at operation %d on %s: %w", i, node.ID, err)
		}
	}
	kvGetDuration := time.Since(start)

	fmt.Printf("✓ Cluster KV Put Performance: %d ops in %v (%.2f ops/sec)\n", operations, kvPutDuration, float64(operations)/kvPutDuration.Seconds())
	fmt.Printf("✓ Cluster KV Get Performance: %d ops in %v (%.2f ops/sec)\n", operations, kvGetDuration, float64(operations)/kvGetDuration.Seconds())

	return nil
}

func testClusterDataConsistency(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing data consistency across cluster...")

	// Create KV store with replication
	kvConfig := hub.KeyValueStoreConfig{
		Bucket:   "cluster_consistency_test",
		Replicas: 3,
	}
	err = nodes[0].Hub.CreateOrUpdateKeyValueStore(kvConfig)
	if err != nil {
		return fmt.Errorf("failed to create KV store: %w", err)
	}

	// Put a value from one node
	testKey := "consistency_test_key"
	testValue := "consistency_test_value"

	_, err = nodes[0].Hub.PutToKeyValueStore("cluster_consistency_test", testKey, []byte(testValue))
	if err != nil {
		return fmt.Errorf("failed to put test value: %w", err)
	}

	// Wait for replication
	time.Sleep(3 * time.Second)

	// Verify the value is consistent across all nodes
	for _, node := range nodes {
		value, _, err := node.Hub.GetFromKeyValueStore("cluster_consistency_test", testKey)
		if err != nil {
			return fmt.Errorf("failed to get value from %s: %w", node.ID, err)
		}

		if string(value) != testValue {
			return fmt.Errorf("data inconsistency on %s: expected %s, got %s", node.ID, testValue, string(value))
		}

		fmt.Printf("✓ Node %s has consistent data: %s\n", node.ID, string(value))
	}

	fmt.Printf("✓ Data consistency verified across all nodes\n")
	return nil
}

func testClusterNetworkResilience(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing network resilience...")

	// Create persistent stream
	config := &hub.PersistentConfig{
		Description: "Network resilience test stream",
		Subjects:    []string{"cluster.resilience.>"},
		Retention:   0,
		MaxMsgs:     100,
		Replicas:    3,
	}

	err = nodes[0].Hub.CreateOrUpdatePersistent(config)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// Publish messages to test persistence
	messageCount := 5
	for i := 0; i < messageCount; i++ {
		msg := []byte(fmt.Sprintf("resilience-message-%d", i+1))
		err := nodes[i%len(nodes)].Hub.PublishPersistent("cluster.resilience.test", msg)
		if err != nil {
			return fmt.Errorf("failed to publish message %d: %w", i+1, err)
		}
	}

	// Subscribe and verify all messages are recoverable
	received := make(chan []byte, messageCount)
	var receivedCount int
	var mu sync.Mutex

	cancel, err := nodes[2].Hub.SubscribePersistentViaDurable("resilience-consumer", "cluster.resilience.>", func(subject string, msg []byte) ([]byte, bool, bool) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		received <- msg
		return nil, false, true
	}, func(err error) {
		log.Printf("Subscription error: %v", err)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	defer cancel()

	// Wait for all messages
	timeout := time.After(10 * time.Second)
	for i := 0; i < messageCount; i++ {
		select {
		case msg := <-received:
			fmt.Printf("✓ Recovered message: %s\n", string(msg))
		case <-timeout:
			return fmt.Errorf("timeout waiting for message %d, received %d", i+1, receivedCount)
		}
	}

	mu.Lock()
	if receivedCount != messageCount {
		return fmt.Errorf("expected %d messages, got %d", messageCount, receivedCount)
	}
	mu.Unlock()

	fmt.Printf("✓ Network resilience test successful (%d messages recovered)\n", receivedCount)
	return nil
}

func testClusterStreamReplication(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing stream replication across cluster...")

	// Create multiple streams with different replication factors
	streams := []struct {
		name     string
		subjects []string
		replicas int
	}{
		{"stream-r1", []string{"test.r1.>"}, 1},
		{"stream-r3", []string{"test.r3.>"}, 3},
	}

	for _, s := range streams {
		config := &hub.PersistentConfig{
			Description: fmt.Sprintf("Test stream %s with %d replicas", s.name, s.replicas),
			Subjects:    s.subjects,
			Retention:   0,
			MaxMsgs:     100,
			Replicas:    s.replicas,
		}

		err = nodes[0].Hub.CreateOrUpdatePersistent(config)
		if err != nil {
			return fmt.Errorf("failed to create stream %s: %w", s.name, err)
		}

		// Test publishing to this stream
		testMsg := []byte(fmt.Sprintf("Test message for %s", s.name))
		err = nodes[1].Hub.PublishPersistent(s.subjects[0], testMsg)
		if err != nil {
			return fmt.Errorf("failed to publish to stream %s: %w", s.name, err)
		}

		fmt.Printf("✓ Created and tested stream %s with %d replicas\n", s.name, s.replicas)
	}

	fmt.Printf("✓ Stream replication test successful\n")
	return nil
}

func testClusterConsumerManagement(tempDir string) error {
	nodeCount := 3
	nodes, err := createCluster(nodeCount, tempDir)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer shutdownCluster(nodes)

	fmt.Println("Testing consumer management across cluster...")

	// Create stream
	config := &hub.PersistentConfig{
		Description:  "Consumer management test stream",
		Subjects:     []string{"cluster.consumer.>"},
		Retention:    0,
		MaxConsumers: 5,
		MaxMsgs:      100,
		Replicas:     3,
	}

	err = nodes[0].Hub.CreateOrUpdatePersistent(config)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// Create consumers on different nodes
	consumerResults := make(map[string]int)
	var mu sync.Mutex
	cancelFuncs := make([]func(), 0)

	for i, node := range nodes {
		consumerID := fmt.Sprintf("cluster-consumer-%d", i+1)

		cancel, err := node.Hub.SubscribePersistentViaDurable(consumerID, "cluster.consumer.events", func(subject string, msg []byte) ([]byte, bool, bool) {
			mu.Lock()
			consumerResults[consumerID]++
			mu.Unlock()
			fmt.Printf("Consumer %s on %s received: %s\n", consumerID, node.ID, string(msg))
			return nil, false, true
		}, func(err error) {
			log.Printf("Consumer %s error: %v", consumerID, err)
		})

		if err != nil {
			for _, c := range cancelFuncs {
				c()
			}
			return fmt.Errorf("failed to create consumer %s: %w", consumerID, err)
		}
		cancelFuncs = append(cancelFuncs, cancel)
	}
	defer func() {
		for _, cancel := range cancelFuncs {
			cancel()
		}
	}()

	// Wait for consumers to be ready
	time.Sleep(2 * time.Second)

	// Publish test messages
	messageCount := 6
	for i := 0; i < messageCount; i++ {
		msg := []byte(fmt.Sprintf("consumer-test-message-%d", i+1))
		err = nodes[i%len(nodes)].Hub.PublishPersistent("cluster.consumer.events", msg)
		if err != nil {
			return fmt.Errorf("failed to publish message %d: %w", i+1, err)
		}
	}

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Verify message distribution
	mu.Lock()
	totalReceived := 0
	for consumerID, count := range consumerResults {
		fmt.Printf("Consumer %s processed %d messages\n", consumerID, count)
		totalReceived += count
	}
	mu.Unlock()

	if totalReceived != messageCount {
		return fmt.Errorf("expected %d messages total, got %d", messageCount, totalReceived)
	}

	fmt.Printf("✓ Consumer management test successful (%d messages distributed)\n", totalReceived)
	return nil
}

func testClusterConfigurationValidation(tempDir string) error {
	fmt.Println("Testing cluster configuration validation...")

	// Test invalid cluster configuration
	invalidOpts := &hub.Options{
		Name:        "invalid-node",
		Host:        "127.0.0.1",
		Port:        14999,
		ClusterHost: "127.0.0.1",
		ClusterPort: 0, // Invalid cluster port
		StoreDir:    filepath.Join(tempDir, "invalid_test"),
	}

	_, err := hub.NewHub(invalidOpts)
	if err == nil {
		return fmt.Errorf("expected error with invalid cluster configuration")
	}
	fmt.Printf("✓ Correctly rejected invalid cluster configuration: %v\n", err)

	// Test valid minimal cluster configuration
	validOpts := &hub.Options{
		Name:                "valid-node",
		Host:                "127.0.0.1",
		Port:                14998,
		ClusterHost:         "127.0.0.1",
		ClusterPort:         16998,
		ClusterConnPoolSize: 4,
		ClusterPingInterval: 30 * time.Second,
		StoreDir:            filepath.Join(tempDir, "valid_test"),
		JetstreamMaxMemory:  hub.NewSizeFromMegabytes(128),
		JetstreamMaxStorage: hub.NewSizeFromMegabytes(256),
	}

	h, err := hub.NewHub(validOpts)
	if err != nil {
		return fmt.Errorf("failed to create hub with valid cluster config: %w", err)
	}
	defer h.Shutdown()

	fmt.Printf("✓ Successfully created hub with valid cluster configuration\n")

	// Test basic functionality
	testMsg := []byte("cluster config validation test")
	err = h.PublishVolatile("test.config", testMsg)
	if err != nil {
		return fmt.Errorf("failed to publish with configured cluster hub: %w", err)
	}

	fmt.Printf("✓ Cluster configuration validation completed successfully\n")
	return nil
}
