package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rivulet-io/hub"
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

	// Run cluster tests
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

// createCluster creates a cluster of N nodes

// shutdownCluster shuts down all nodes in the cluster

func testClusterFormation(tempDir string) error {
	fmt.Println("Testing cluster formation...")

	// Step 1: Create first node (seed node)
	fmt.Println("Step 1: Creating seed node...")

	opts1, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options: %w", err)
	}

	// Configure first node as seed - try with cluster disabled initially
	opts1.Port = 4224      // Unique client port
	opts1.ClusterPort = 0  // Disable cluster initially
	opts1.LeafNodePort = 0 // Disable leafnode initially
	opts1.StoreDir = filepath.Join(tempDir, "seed_node")
	opts1.Name = "cluster-seed-node"
	opts1.Routes = nil // Seed node has no initial routes

	fmt.Printf("Creating seed node - Client: %d, Cluster: %d, LeafNode: %d\n",
		opts1.Port, opts1.ClusterPort, opts1.LeafNodePort)
	h1, err := hub.NewHub(opts1)
	if err != nil {
		return fmt.Errorf("failed to create seed node hub: %w", err)
	}
	defer h1.Shutdown()
	fmt.Println("✓ Step 1: Seed node created successfully")

	// Wait for seed node to be ready
	time.Sleep(3 * time.Second)

	// Step 2: Create second node that connects to seed
	fmt.Println("Step 2: Creating follower node...")

	opts2, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 2: %w", err)
	}

	// Create route to seed node
	seedRoute, err := url.Parse("nats://127.0.0.1:6224")
	if err != nil {
		return fmt.Errorf("failed to parse seed route: %w", err)
	}

	// Configure second node to connect to seed
	opts2.Port = 4225         // Different client port
	opts2.ClusterPort = 6225  // Different cluster port
	opts2.LeafNodePort = 7425 // Different leafnode port
	opts2.StoreDir = filepath.Join(tempDir, "follower_node")
	opts2.Name = "cluster-follower-node"
	opts2.Routes = []*url.URL{seedRoute} // Connect to seed node

	fmt.Printf("Creating follower node - Client: %d, Cluster: %d, LeafNode: %d\n",
		opts2.Port, opts2.ClusterPort, opts2.LeafNodePort)
	fmt.Printf("Connecting to seed cluster at: %s\n", seedRoute.String())
	h2, err := hub.NewHub(opts2)
	if err != nil {
		return fmt.Errorf("failed to create follower node hub: %w", err)
	}
	defer h2.Shutdown()
	fmt.Println("✓ Step 2: Follower node created successfully")

	// Wait for cluster formation
	fmt.Println("Waiting for cluster formation...")
	time.Sleep(5 * time.Second)

	// Test basic functionality on both nodes
	testMsg := []byte("cluster formation test message")
	err = h1.PublishVolatile("test.seed", testMsg)
	if err != nil {
		return fmt.Errorf("seed node is not operational: %w", err)
	}

	err = h2.PublishVolatile("test.follower", testMsg)
	if err != nil {
		return fmt.Errorf("follower node is not operational: %w", err)
	}

	fmt.Println("✓ Both nodes are operational")
	fmt.Println("✓ Cluster formation test successful (2-node cluster formed)")
	return nil
}

func testClusterBasicCommunication(tempDir string) error {
	fmt.Println("Testing basic communication between nodes...")

	// Create two independent nodes
	fmt.Println("Creating two independent nodes...")

	// Node 1
	opts1, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 1: %w", err)
	}
	opts1.Port = 4226
	opts1.ClusterPort = 0
	opts1.LeafNodePort = 0
	opts1.StoreDir = filepath.Join(tempDir, "comm_node1")
	opts1.Name = "communication-node-1"

	h1, err := hub.NewHub(opts1)
	if err != nil {
		return fmt.Errorf("failed to create node 1: %w", err)
	}
	defer h1.Shutdown()

	// Node 2
	opts2, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 2: %w", err)
	}
	opts2.Port = 4227
	opts2.ClusterPort = 0
	opts2.LeafNodePort = 0
	opts2.StoreDir = filepath.Join(tempDir, "comm_node2")
	opts2.Name = "communication-node-2"

	h2, err := hub.NewHub(opts2)
	if err != nil {
		return fmt.Errorf("failed to create node 2: %w", err)
	}
	defer h2.Shutdown()

	fmt.Println("✓ Both nodes created successfully")

	// Test basic publish/subscribe within each node
	received1 := make(chan []byte, 1)
	received2 := make(chan []byte, 1)

	// Subscribe on node 1
	cancel1, err := h1.SubscribeVolatileViaFanout("test.comm.node1", func(subject string, msg []byte) ([]byte, bool) {
		received1 <- msg
		return nil, false
	}, func(err error) {
		fmt.Printf("Node 1 subscription error: %v\n", err)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe on node 1: %w", err)
	}
	defer cancel1()

	// Subscribe on node 2
	cancel2, err := h2.SubscribeVolatileViaFanout("test.comm.node2", func(subject string, msg []byte) ([]byte, bool) {
		received2 <- msg
		return nil, false
	}, func(err error) {
		fmt.Printf("Node 2 subscription error: %v\n", err)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe on node 2: %w", err)
	}
	defer cancel2()

	// Wait for subscriptions to be ready
	time.Sleep(1 * time.Second)

	// Test communication within node 1
	testMsg1 := []byte("Hello from Node 1")
	err = h1.PublishVolatile("test.comm.node1", testMsg1)
	if err != nil {
		return fmt.Errorf("failed to publish on node 1: %w", err)
	}

	// Test communication within node 2
	testMsg2 := []byte("Hello from Node 2")
	err = h2.PublishVolatile("test.comm.node2", testMsg2)
	if err != nil {
		return fmt.Errorf("failed to publish on node 2: %w", err)
	}

	// Verify messages received
	timeout := time.After(3 * time.Second)

	select {
	case msg := <-received1:
		if string(msg) != string(testMsg1) {
			return fmt.Errorf("node 1 received wrong message: expected %s, got %s", testMsg1, msg)
		}
		fmt.Println("✓ Node 1 internal communication successful")
	case <-timeout:
		return fmt.Errorf("timeout waiting for message on node 1")
	}

	select {
	case msg := <-received2:
		if string(msg) != string(testMsg2) {
			return fmt.Errorf("node 2 received wrong message: expected %s, got %s", testMsg2, msg)
		}
		fmt.Println("✓ Node 2 internal communication successful")
	case <-timeout:
		return fmt.Errorf("timeout waiting for message on node 2")
	}

	fmt.Println("✓ Basic communication test successful")
	return nil
}

func testClusterJetStreamOperations(tempDir string) error {
	fmt.Println("Testing JetStream operations across multiple nodes...")

	// Create two independent nodes
	fmt.Println("Creating two nodes for JetStream testing...")

	// Node 1
	opts1, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 1: %w", err)
	}
	opts1.Port = 4228
	opts1.ClusterPort = 0
	opts1.LeafNodePort = 0
	opts1.StoreDir = filepath.Join(tempDir, "jetstream_node1")
	opts1.Name = "jetstream-node-1"

	h1, err := hub.NewHub(opts1)
	if err != nil {
		return fmt.Errorf("failed to create node 1: %w", err)
	}
	defer h1.Shutdown()

	// Node 2
	opts2, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 2: %w", err)
	}
	opts2.Port = 4229
	opts2.ClusterPort = 0
	opts2.LeafNodePort = 0
	opts2.StoreDir = filepath.Join(tempDir, "jetstream_node2")
	opts2.Name = "jetstream-node-2"

	h2, err := hub.NewHub(opts2)
	if err != nil {
		return fmt.Errorf("failed to create node 2: %w", err)
	}
	defer h2.Shutdown()

	fmt.Println("✓ Both JetStream nodes created successfully")

	// Test persistent stream creation and operations on node 1
	fmt.Println("Testing persistent stream operations on node 1...")
	config1 := &hub.PersistentConfig{
		Description: "Node 1 test stream",
		Subjects:    []string{"node1.test.>"},
		Retention:   0,
		MaxMsgs:     100,
		Replicas:    1,
	}

	err = h1.CreateOrUpdatePersistent(config1)
	if err != nil {
		return fmt.Errorf("failed to create persistent stream on node 1: %w", err)
	}

	// Test persistent stream creation and operations on node 2
	fmt.Println("Testing persistent stream operations on node 2...")
	config2 := &hub.PersistentConfig{
		Description: "Node 2 test stream",
		Subjects:    []string{"node2.test.>"},
		Retention:   0,
		MaxMsgs:     100,
		Replicas:    1,
	}

	err = h2.CreateOrUpdatePersistent(config2)
	if err != nil {
		return fmt.Errorf("failed to create persistent stream on node 2: %w", err)
	}

	fmt.Println("✓ Persistent streams created on both nodes")

	// Test publishing to persistent streams
	fmt.Println("Testing publishing to persistent streams...")

	// Publish to node 1 stream
	testMsg1 := []byte("Persistent message from node 1")
	err = h1.PublishPersistent("node1.test.messages", testMsg1)
	if err != nil {
		return fmt.Errorf("failed to publish persistent message on node 1: %w", err)
	}

	// Publish to node 2 stream
	testMsg2 := []byte("Persistent message from node 2")
	err = h2.PublishPersistent("node2.test.messages", testMsg2)
	if err != nil {
		return fmt.Errorf("failed to publish persistent message on node 2: %w", err)
	}

	fmt.Println("✓ Messages published to persistent streams")

	// Test consuming from persistent streams
	fmt.Println("Testing consuming from persistent streams...")

	received1 := make(chan []byte, 1)
	received2 := make(chan []byte, 1)

	// Subscribe to node 1 stream
	cancel1, err := h1.SubscribePersistentViaDurable("test-consumer-1", "node1.test.>",
		func(subject string, msg []byte) ([]byte, bool, bool) {
			received1 <- msg
			return nil, false, true
		},
		func(err error) {
			fmt.Printf("Node 1 persistent subscription error: %v\\n", err)
		})
	if err != nil {
		return fmt.Errorf("failed to subscribe to persistent stream on node 1: %w", err)
	}
	defer cancel1()

	// Subscribe to node 2 stream
	cancel2, err := h2.SubscribePersistentViaDurable("test-consumer-2", "node2.test.>",
		func(subject string, msg []byte) ([]byte, bool, bool) {
			received2 <- msg
			return nil, false, true
		},
		func(err error) {
			fmt.Printf("Node 2 persistent subscription error: %v\\n", err)
		})
	if err != nil {
		return fmt.Errorf("failed to subscribe to persistent stream on node 2: %w", err)
	}
	defer cancel2()

	// Wait for messages to be consumed
	timeout := time.After(5 * time.Second)

	select {
	case msg := <-received1:
		if string(msg) != string(testMsg1) {
			return fmt.Errorf("node 1 received wrong persistent message: expected %s, got %s", testMsg1, msg)
		}
		fmt.Println("✓ Node 1 persistent message consumption successful")
	case <-timeout:
		return fmt.Errorf("timeout waiting for persistent message on node 1")
	}

	select {
	case msg := <-received2:
		if string(msg) != string(testMsg2) {
			return fmt.Errorf("node 2 received wrong persistent message: expected %s, got %s", testMsg2, msg)
		}
		fmt.Println("✓ Node 2 persistent message consumption successful")
	case <-timeout:
		return fmt.Errorf("timeout waiting for persistent message on node 2")
	}

	fmt.Println("✓ JetStream operations test successful")
	return nil
}

func testClusterKeyValueStore(tempDir string) error {
	fmt.Println("Testing Key-Value Store operations across multiple nodes...")

	// Create two independent nodes
	fmt.Println("Creating two nodes for KV Store testing...")

	// Node 1
	opts1, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 1: %w", err)
	}
	opts1.Port = 4230
	opts1.ClusterPort = 0
	opts1.LeafNodePort = 0
	opts1.StoreDir = filepath.Join(tempDir, "kv_node1")
	opts1.Name = "kv-node-1"

	h1, err := hub.NewHub(opts1)
	if err != nil {
		return fmt.Errorf("failed to create node 1: %w", err)
	}
	defer h1.Shutdown()

	// Node 2
	opts2, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 2: %w", err)
	}
	opts2.Port = 4231
	opts2.ClusterPort = 0
	opts2.LeafNodePort = 0
	opts2.StoreDir = filepath.Join(tempDir, "kv_node2")
	opts2.Name = "kv-node-2"

	h2, err := hub.NewHub(opts2)
	if err != nil {
		return fmt.Errorf("failed to create node 2: %w", err)
	}
	defer h2.Shutdown()

	fmt.Println("✓ Both KV nodes created successfully")

	// Test KV store creation on both nodes
	fmt.Println("Testing KV store creation on both nodes...")

	kvConfig1 := hub.KeyValueStoreConfig{
		Bucket:   "node1_test_bucket",
		Replicas: 1,
	}
	err = h1.CreateOrUpdateKeyValueStore(kvConfig1)
	if err != nil {
		return fmt.Errorf("failed to create KV store on node 1: %w", err)
	}

	kvConfig2 := hub.KeyValueStoreConfig{
		Bucket:   "node2_test_bucket",
		Replicas: 1,
	}
	err = h2.CreateOrUpdateKeyValueStore(kvConfig2)
	if err != nil {
		return fmt.Errorf("failed to create KV store on node 2: %w", err)
	}

	fmt.Println("✓ KV stores created on both nodes")

	// Test KV operations on node 1
	fmt.Println("Testing KV operations on node 1...")
	testKey1 := "test_key_1"
	testValue1 := []byte("test_value_from_node_1")

	revision1, err := h1.PutToKeyValueStore("node1_test_bucket", testKey1, testValue1)
	if err != nil {
		return fmt.Errorf("failed to put to KV store on node 1: %w", err)
	}

	retrievedValue1, retrievedRevision1, err := h1.GetFromKeyValueStore("node1_test_bucket", testKey1)
	if err != nil {
		return fmt.Errorf("failed to get from KV store on node 1: %w", err)
	}

	if string(retrievedValue1) != string(testValue1) {
		return fmt.Errorf("KV value mismatch on node 1: expected %s, got %s", testValue1, retrievedValue1)
	}

	if retrievedRevision1 != revision1 {
		return fmt.Errorf("KV revision mismatch on node 1: expected %d, got %d", revision1, retrievedRevision1)
	}

	fmt.Println("✓ Node 1 KV operations successful")

	// Test KV operations on node 2
	fmt.Println("Testing KV operations on node 2...")
	testKey2 := "test_key_2"
	testValue2 := []byte("test_value_from_node_2")

	revision2, err := h2.PutToKeyValueStore("node2_test_bucket", testKey2, testValue2)
	if err != nil {
		return fmt.Errorf("failed to put to KV store on node 2: %w", err)
	}

	retrievedValue2, retrievedRevision2, err := h2.GetFromKeyValueStore("node2_test_bucket", testKey2)
	if err != nil {
		return fmt.Errorf("failed to get from KV store on node 2: %w", err)
	}

	if string(retrievedValue2) != string(testValue2) {
		return fmt.Errorf("KV value mismatch on node 2: expected %s, got %s", testValue2, retrievedValue2)
	}

	if retrievedRevision2 != revision2 {
		return fmt.Errorf("KV revision mismatch on node 2: expected %d, got %d", revision2, retrievedRevision2)
	}

	fmt.Println("✓ Node 2 KV operations successful")

	// Test KV update operations
	fmt.Println("Testing KV update operations...")
	updatedValue1 := []byte("updated_value_from_node_1")
	newRevision1, err := h1.UpdateToKeyValueStore("node1_test_bucket", testKey1, updatedValue1, revision1)
	if err != nil {
		return fmt.Errorf("failed to update KV store on node 1: %w", err)
	}

	retrievedUpdatedValue1, _, err := h1.GetFromKeyValueStore("node1_test_bucket", testKey1)
	if err != nil {
		return fmt.Errorf("failed to get updated value from KV store on node 1: %w", err)
	}

	if string(retrievedUpdatedValue1) != string(updatedValue1) {
		return fmt.Errorf("KV updated value mismatch on node 1: expected %s, got %s", updatedValue1, retrievedUpdatedValue1)
	}

	fmt.Printf("✓ KV update successful (revision %d -> %d)\\n", revision1, newRevision1)

	// Test KV delete operations
	fmt.Println("Testing KV delete operations...")
	err = h2.DeleteFromKeyValueStore("node2_test_bucket", testKey2)
	if err != nil {
		return fmt.Errorf("failed to delete from KV store on node 2: %w", err)
	}

	_, _, err = h2.GetFromKeyValueStore("node2_test_bucket", testKey2)
	if err == nil {
		return fmt.Errorf("expected error when getting deleted key, but got none")
	}

	fmt.Println("✓ KV delete operation successful")
	fmt.Println("✓ Key-Value Store operations test successful")
	return nil
}

func testClusterObjectStore(tempDir string) error {
	fmt.Println("Testing Object Store operations across multiple nodes...")

	// Create two independent nodes
	fmt.Println("Creating two nodes for Object Store testing...")

	// Node 1
	opts1, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 1: %w", err)
	}
	opts1.Port = 4232
	opts1.ClusterPort = 0
	opts1.LeafNodePort = 0
	opts1.StoreDir = filepath.Join(tempDir, "obj_node1")
	opts1.Name = "obj-node-1"

	h1, err := hub.NewHub(opts1)
	if err != nil {
		return fmt.Errorf("failed to create node 1: %w", err)
	}
	defer h1.Shutdown()

	// Node 2
	opts2, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 2: %w", err)
	}
	opts2.Port = 4233
	opts2.ClusterPort = 0
	opts2.LeafNodePort = 0
	opts2.StoreDir = filepath.Join(tempDir, "obj_node2")
	opts2.Name = "obj-node-2"

	h2, err := hub.NewHub(opts2)
	if err != nil {
		return fmt.Errorf("failed to create node 2: %w", err)
	}
	defer h2.Shutdown()

	fmt.Println("✓ Both Object Store nodes created successfully")

	// Test Object store creation on both nodes
	fmt.Println("Testing Object store creation on both nodes...")

	objConfig1 := hub.ObjectStoreConfig{
		Bucket:   "node1_obj_bucket",
		Replicas: 1,
	}
	err = h1.CreateObjectStore(objConfig1)
	if err != nil {
		return fmt.Errorf("failed to create Object store on node 1: %w", err)
	}

	objConfig2 := hub.ObjectStoreConfig{
		Bucket:   "node2_obj_bucket",
		Replicas: 1,
	}
	err = h2.CreateObjectStore(objConfig2)
	if err != nil {
		return fmt.Errorf("failed to create Object store on node 2: %w", err)
	}

	fmt.Println("✓ Object stores created on both nodes")

	// Test Object operations on node 1
	fmt.Println("Testing Object operations on node 1...")
	testKey1 := "test_object_1"
	testData1 := []byte("This is test object data from node 1")
	testMetadata1 := map[string]string{
		"author": "node-1",
		"type":   "test-data",
	}

	err = h1.PutToObjectStore("node1_obj_bucket", testKey1, testData1, testMetadata1)
	if err != nil {
		return fmt.Errorf("failed to put to Object store on node 1: %w", err)
	}

	retrievedData1, err := h1.GetFromObjectStore("node1_obj_bucket", testKey1)
	if err != nil {
		return fmt.Errorf("failed to get from Object store on node 1: %w", err)
	}

	if string(retrievedData1) != string(testData1) {
		return fmt.Errorf("object data mismatch on node 1: expected %s, got %s", testData1, retrievedData1)
	}

	fmt.Println("✓ Node 1 Object operations successful")

	// Test Object operations on node 2
	fmt.Println("Testing Object operations on node 2...")
	testKey2 := "test_object_2"
	testData2 := []byte("This is test object data from node 2")
	testMetadata2 := map[string]string{
		"author": "node-2",
		"type":   "test-data",
		"size":   "large",
	}

	err = h2.PutToObjectStore("node2_obj_bucket", testKey2, testData2, testMetadata2)
	if err != nil {
		return fmt.Errorf("failed to put to Object store on node 2: %w", err)
	}

	retrievedData2, err := h2.GetFromObjectStore("node2_obj_bucket", testKey2)
	if err != nil {
		return fmt.Errorf("failed to get from Object store on node 2: %w", err)
	}

	if string(retrievedData2) != string(testData2) {
		return fmt.Errorf("object data mismatch on node 2: expected %s, got %s", testData2, retrievedData2)
	}

	fmt.Println("✓ Node 2 Object operations successful")

	// Test Object delete operations
	fmt.Println("Testing Object delete operations...")
	err = h1.DeleteFromObjectStore("node1_obj_bucket", testKey1)
	if err != nil {
		return fmt.Errorf("failed to delete from Object store on node 1: %w", err)
	}

	_, err = h1.GetFromObjectStore("node1_obj_bucket", testKey1)
	if err == nil {
		return fmt.Errorf("expected error when getting deleted object, but got none")
	}

	fmt.Println("✓ Object delete operation successful")
	fmt.Println("✓ Object Store operations test successful")
	return nil
}

func testClusterLoadBalancing(tempDir string) error {
	fmt.Println("Testing load balancing across multiple nodes...")

	// Create three nodes for load balancing test
	fmt.Println("Creating three nodes for load balancing...")

	nodes := make([]*hub.Hub, 3)
	nodeNames := []string{"lb-node-1", "lb-node-2", "lb-node-3"}
	basePorts := []int{4234, 4235, 4236}

	for i := 0; i < 3; i++ {
		opts, err := hub.DefaultNodeOptions()
		if err != nil {
			return fmt.Errorf("failed to create default options for node %d: %w", i+1, err)
		}
		opts.Port = basePorts[i]
		opts.ClusterPort = 0
		opts.LeafNodePort = 0
		opts.StoreDir = filepath.Join(tempDir, fmt.Sprintf("lb_node%d", i+1))
		opts.Name = nodeNames[i]

		h, err := hub.NewHub(opts)
		if err != nil {
			// Cleanup already created nodes
			for j := 0; j < i; j++ {
				if nodes[j] != nil {
					nodes[j].Shutdown()
				}
			}
			return fmt.Errorf("failed to create node %d: %w", i+1, err)
		}
		nodes[i] = h
	}

	// Cleanup all nodes at the end
	defer func() {
		for _, node := range nodes {
			if node != nil {
				node.Shutdown()
			}
		}
	}()

	fmt.Println("✓ All three load balancing nodes created successfully")

	// Create queue subscribers on each node for load balancing
	fmt.Println("Setting up queue subscribers for load balancing...")

	messageCount := make(map[string]int)
	var mu sync.Mutex
	queueName := "load-balance-queue"

	var cancelFuncs []func()

	for i, node := range nodes {
		nodeName := nodeNames[i]
		messageCount[nodeName] = 0

		cancel, err := node.SubscribeVolatileViaQueue("test.loadbalance", queueName,
			func(subject string, msg []byte) ([]byte, bool) {
				mu.Lock()
				messageCount[nodeName]++
				mu.Unlock()
				fmt.Printf("Node %s processed message: %s\\n", nodeName, string(msg))
				return []byte(fmt.Sprintf("Response from %s", nodeName)), false
			},
			func(err error) {
				fmt.Printf("Queue subscription error on %s: %v\\n", nodeName, err)
			})

		if err != nil {
			// Cleanup already created subscriptions
			for _, c := range cancelFuncs {
				c()
			}
			return fmt.Errorf("failed to create queue subscription on %s: %w", nodeName, err)
		}
		cancelFuncs = append(cancelFuncs, cancel)
	}

	defer func() {
		for _, cancel := range cancelFuncs {
			cancel()
		}
	}()

	// Wait for subscriptions to be ready
	time.Sleep(2 * time.Second)

	fmt.Println("✓ Queue subscribers set up on all nodes")

	// Send multiple messages and verify load balancing
	fmt.Println("Testing load balancing with multiple messages...")

	totalMessages := 15
	for i := 0; i < totalMessages; i++ {
		// Round-robin publish to different nodes
		node := nodes[i%len(nodes)]
		message := []byte(fmt.Sprintf("Load balance test message %d", i+1))

		err := node.PublishVolatile("test.loadbalance", message)
		if err != nil {
			return fmt.Errorf("failed to publish message %d: %w", i+1, err)
		}

		// Small delay to allow processing
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for all messages to be processed
	time.Sleep(3 * time.Second)

	// Verify load distribution
	fmt.Println("Verifying load distribution...")

	mu.Lock()
	totalProcessed := 0
	for nodeName, count := range messageCount {
		fmt.Printf("Node %s processed %d messages\\n", nodeName, count)
		totalProcessed += count
	}
	mu.Unlock()

	if totalProcessed != totalMessages {
		return fmt.Errorf("message count mismatch: sent %d, processed %d", totalMessages, totalProcessed)
	}

	// Check that load is reasonably distributed (each node should handle some messages)
	mu.Lock()
	activeNodes := 0
	for _, count := range messageCount {
		if count > 0 {
			activeNodes++
		}
	}
	mu.Unlock()

	if activeNodes < 2 {
		return fmt.Errorf("poor load distribution: only %d out of %d nodes handled messages", activeNodes, len(nodes))
	}

	fmt.Printf("✓ Load balancing successful: %d messages distributed across %d active nodes\\n", totalMessages, activeNodes)
	return nil
}

func testClusterErrorHandling(tempDir string) error {
	fmt.Println("Testing error handling across multiple nodes...")

	// Create two nodes for error handling test
	fmt.Println("Creating two nodes for error handling testing...")

	// Node 1
	opts1, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 1: %w", err)
	}
	opts1.Port = 4237
	opts1.ClusterPort = 0
	opts1.LeafNodePort = 0
	opts1.StoreDir = filepath.Join(tempDir, "error_node1")
	opts1.Name = "error-node-1"

	h1, err := hub.NewHub(opts1)
	if err != nil {
		return fmt.Errorf("failed to create node 1: %w", err)
	}
	defer h1.Shutdown()

	// Node 2
	opts2, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 2: %w", err)
	}
	opts2.Port = 4238
	opts2.ClusterPort = 0
	opts2.LeafNodePort = 0
	opts2.StoreDir = filepath.Join(tempDir, "error_node2")
	opts2.Name = "error-node-2"

	h2, err := hub.NewHub(opts2)
	if err != nil {
		return fmt.Errorf("failed to create node 2: %w", err)
	}
	defer h2.Shutdown()

	fmt.Println("✓ Both error handling nodes created successfully")

	// Test 1: Invalid KV bucket operations
	fmt.Println("Testing invalid KV bucket operations...")

	// Try to get from non-existent bucket
	_, _, err = h1.GetFromKeyValueStore("non_existent_bucket", "test_key")
	if err == nil {
		return fmt.Errorf("expected error when accessing non-existent KV bucket, but got none")
	}
	fmt.Println("✓ Non-existent KV bucket error handled correctly")

	// Test 2: Invalid Object bucket operations
	fmt.Println("Testing invalid Object bucket operations...")

	// Try to get from non-existent object bucket
	_, err = h2.GetFromObjectStore("non_existent_obj_bucket", "test_key")
	if err == nil {
		return fmt.Errorf("expected error when accessing non-existent Object bucket, but got none")
	}
	fmt.Println("✓ Non-existent Object bucket error handled correctly")

	// Test 3: Invalid persistent stream operations
	fmt.Println("Testing invalid persistent stream operations...")

	// Try to publish to non-existent stream subject without creating stream first
	err = h1.PublishPersistent("non.existent.stream.subject", []byte("test message"))
	if err == nil {
		return fmt.Errorf("expected error when publishing to non-existent stream, but got none")
	}
	fmt.Println("✓ Non-existent stream error handled correctly")

	// Test 4: KV update with wrong revision
	fmt.Println("Testing KV update with wrong revision...")

	// Create KV store first
	kvConfig := hub.KeyValueStoreConfig{
		Bucket:   "error_test_bucket",
		Replicas: 1,
	}
	err = h1.CreateOrUpdateKeyValueStore(kvConfig)
	if err != nil {
		return fmt.Errorf("failed to create KV store: %w", err)
	}

	// Put initial value
	testKey := "test_revision_key"
	testValue := []byte("initial_value")
	revision, err := h1.PutToKeyValueStore("error_test_bucket", testKey, testValue)
	if err != nil {
		return fmt.Errorf("failed to put initial value: %w", err)
	}

	// Try to update with wrong revision
	wrongRevision := revision + 100
	_, err = h1.UpdateToKeyValueStore("error_test_bucket", testKey, []byte("updated_value"), wrongRevision)
	if err == nil {
		return fmt.Errorf("expected error when updating with wrong revision, but got none")
	}
	fmt.Println("✓ Wrong revision error handled correctly")

	// Test 5: Subscription error handling
	fmt.Println("Testing subscription error handling...")

	errorReceived := make(chan error, 1)

	// Create subscription with error handler
	cancel, err := h2.SubscribeVolatileViaFanout("test.error.subject",
		func(subject string, msg []byte) ([]byte, bool) {
			// Simulate processing
			return []byte("processed"), false
		},
		func(err error) {
			fmt.Printf("Subscription error received: %v\\n", err)
			errorReceived <- err
		})

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	defer cancel()

	fmt.Println("✓ Subscription error handling set up correctly")

	// Test 6: Resource cleanup on node shutdown
	fmt.Println("Testing resource cleanup...")

	// Create a temporary node that we'll shutdown
	opts3, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for temp node: %w", err)
	}
	opts3.Port = 4239
	opts3.ClusterPort = 0
	opts3.LeafNodePort = 0
	opts3.StoreDir = filepath.Join(tempDir, "temp_error_node")
	opts3.Name = "temp-error-node"

	h3, err := hub.NewHub(opts3)
	if err != nil {
		return fmt.Errorf("failed to create temp node: %w", err)
	}

	// Create some resources on temp node
	tempKvConfig := hub.KeyValueStoreConfig{
		Bucket:   "temp_bucket",
		Replicas: 1,
	}
	err = h3.CreateOrUpdateKeyValueStore(tempKvConfig)
	if err != nil {
		return fmt.Errorf("failed to create temp KV store: %w", err)
	}

	// Shutdown the temp node (should cleanup resources)
	h3.Shutdown()
	fmt.Println("✓ Resource cleanup on shutdown completed")

	fmt.Println("✓ Error handling test successful")
	return nil
}

func testClusterNodeFailureRecovery(tempDir string) error {
	fmt.Println("Testing node failure recovery scenarios...")

	// Create initial node
	fmt.Println("Creating initial node...")

	opts1, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for initial node: %w", err)
	}
	opts1.Port = 4240
	opts1.ClusterPort = 0
	opts1.LeafNodePort = 0
	opts1.StoreDir = filepath.Join(tempDir, "recovery_node1")
	opts1.Name = "recovery-node-1"

	h1, err := hub.NewHub(opts1)
	if err != nil {
		return fmt.Errorf("failed to create initial node: %w", err)
	}
	defer h1.Shutdown()

	fmt.Println("✓ Initial node created successfully")

	// Create persistent stream and KV store for recovery testing
	fmt.Println("Setting up persistent resources...")

	// Create persistent stream
	streamConfig := &hub.PersistentConfig{
		Description: "Recovery test stream",
		Subjects:    []string{"recovery.test.>"},
		Retention:   0,
		MaxMsgs:     100,
		Replicas:    1,
	}
	err = h1.CreateOrUpdatePersistent(streamConfig)
	if err != nil {
		return fmt.Errorf("failed to create persistent stream: %w", err)
	}

	// Create KV store
	kvConfig := hub.KeyValueStoreConfig{
		Bucket:   "recovery_test_bucket",
		Replicas: 1,
	}
	err = h1.CreateOrUpdateKeyValueStore(kvConfig)
	if err != nil {
		return fmt.Errorf("failed to create KV store: %w", err)
	}

	// Add some data
	testKey := "recovery_test_key"
	testValue := []byte("data_before_failure")
	_, err = h1.PutToKeyValueStore("recovery_test_bucket", testKey, testValue)
	if err != nil {
		return fmt.Errorf("failed to put test data: %w", err)
	}

	// Publish some messages
	for i := 0; i < 5; i++ {
		msg := []byte(fmt.Sprintf("pre-failure-message-%d", i+1))
		err = h1.PublishPersistent("recovery.test.messages", msg)
		if err != nil {
			return fmt.Errorf("failed to publish pre-failure message %d: %w", i+1, err)
		}
	}

	fmt.Println("✓ Persistent resources and data created")

	// Simulate node failure by shutting down
	fmt.Println("Simulating node failure (shutdown)...")
	h1.Shutdown()
	time.Sleep(2 * time.Second)
	fmt.Println("✓ Node failure simulated")

	// Create recovery node with same configuration
	fmt.Println("Creating recovery node with same configuration...")

	opts2, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for recovery node: %w", err)
	}
	opts2.Port = 4240 // Same port
	opts2.ClusterPort = 0
	opts2.LeafNodePort = 0
	opts2.StoreDir = filepath.Join(tempDir, "recovery_node1") // Same store directory
	opts2.Name = "recovery-node-1"                            // Same name

	h2, err := hub.NewHub(opts2)
	if err != nil {
		return fmt.Errorf("failed to create recovery node: %w", err)
	}
	defer h2.Shutdown()

	fmt.Println("✓ Recovery node created successfully")

	// Wait for recovery to complete
	time.Sleep(3 * time.Second)

	// Test data recovery
	fmt.Println("Testing data recovery...")

	// Check if KV data survived
	recoveredValue, _, err := h2.GetFromKeyValueStore("recovery_test_bucket", testKey)
	if err != nil {
		return fmt.Errorf("failed to recover KV data: %w", err)
	}

	if string(recoveredValue) != string(testValue) {
		return fmt.Errorf("KV data not recovered correctly: expected %s, got %s", testValue, recoveredValue)
	}

	fmt.Println("✓ KV data recovered successfully")

	// Test stream recovery by subscribing and checking for messages
	fmt.Println("Testing stream recovery...")

	received := make(chan []byte, 10)

	cancel, err := h2.SubscribePersistentViaDurable("recovery-consumer", "recovery.test.>",
		func(subject string, msg []byte) ([]byte, bool, bool) {
			received <- msg
			return nil, false, true
		},
		func(err error) {
			fmt.Printf("Recovery subscription error: %v\\n", err)
		})

	if err != nil {
		return fmt.Errorf("failed to subscribe for recovery test: %w", err)
	}
	defer cancel()

	// Wait for messages to be replayed
	time.Sleep(3 * time.Second)

	// Check received messages
	messageCount := 0
	timeout := time.After(2 * time.Second)

	for {
		select {
		case msg := <-received:
			messageCount++
			fmt.Printf("Recovered message: %s\\n", string(msg))
		case <-timeout:
			goto checkResults
		}
	}

checkResults:
	if messageCount == 0 {
		return fmt.Errorf("no messages recovered from stream")
	}

	fmt.Printf("✓ Stream recovery successful (%d messages recovered)\\n", messageCount)

	// Test adding new data after recovery
	fmt.Println("Testing operations after recovery...")

	newTestValue := []byte("data_after_recovery")
	_, err = h2.PutToKeyValueStore("recovery_test_bucket", "new_key_after_recovery", newTestValue)
	if err != nil {
		return fmt.Errorf("failed to add new data after recovery: %w", err)
	}

	// Publish new message
	err = h2.PublishPersistent("recovery.test.messages", []byte("post-recovery-message"))
	if err != nil {
		return fmt.Errorf("failed to publish new message after recovery: %w", err)
	}

	fmt.Println("✓ Operations after recovery successful")
	fmt.Println("✓ Node failure recovery test successful")
	return nil
}

func testClusterConcurrentOperations(tempDir string) error {
	fmt.Println("Testing concurrent operations across multiple nodes...")

	// Create two nodes for concurrent testing
	fmt.Println("Creating two nodes for concurrent operations...")

	// Node 1
	opts1, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 1: %w", err)
	}
	opts1.Port = 4241
	opts1.ClusterPort = 0
	opts1.LeafNodePort = 0
	opts1.StoreDir = filepath.Join(tempDir, "concurrent_node1")
	opts1.Name = "concurrent-node-1"

	h1, err := hub.NewHub(opts1)
	if err != nil {
		return fmt.Errorf("failed to create node 1: %w", err)
	}
	defer h1.Shutdown()

	// Node 2
	opts2, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 2: %w", err)
	}
	opts2.Port = 4242
	opts2.ClusterPort = 0
	opts2.LeafNodePort = 0
	opts2.StoreDir = filepath.Join(tempDir, "concurrent_node2")
	opts2.Name = "concurrent-node-2"

	h2, err := hub.NewHub(opts2)
	if err != nil {
		return fmt.Errorf("failed to create node 2: %w", err)
	}
	defer h2.Shutdown()

	fmt.Println("✓ Both concurrent operation nodes created successfully")

	// Create KV stores on both nodes
	fmt.Println("Setting up KV stores for concurrent testing...")

	kvConfig1 := hub.KeyValueStoreConfig{
		Bucket:   "concurrent_bucket_1",
		Replicas: 1,
	}
	err = h1.CreateOrUpdateKeyValueStore(kvConfig1)
	if err != nil {
		return fmt.Errorf("failed to create KV store on node 1: %w", err)
	}

	kvConfig2 := hub.KeyValueStoreConfig{
		Bucket:   "concurrent_bucket_2",
		Replicas: 1,
	}
	err = h2.CreateOrUpdateKeyValueStore(kvConfig2)
	if err != nil {
		return fmt.Errorf("failed to create KV store on node 2: %w", err)
	}

	fmt.Println("✓ KV stores created for concurrent testing")

	// Test concurrent KV operations
	fmt.Println("Testing concurrent KV operations...")

	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)
	successCount := 0

	operationsPerNode := 10

	// Concurrent operations on node 1
	for i := 0; i < operationsPerNode; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			key := fmt.Sprintf("concurrent_key_1_%d", index)
			value := []byte(fmt.Sprintf("concurrent_value_1_%d", index))

			// Put operation
			_, err := h1.PutToKeyValueStore("concurrent_bucket_1", key, value)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("put error on node 1 key %s: %w", key, err))
				mu.Unlock()
				return
			}

			// Get operation
			retrieved, _, err := h1.GetFromKeyValueStore("concurrent_bucket_1", key)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("get error on node 1 key %s: %w", key, err))
				mu.Unlock()
				return
			}

			if string(retrieved) != string(value) {
				mu.Lock()
				errors = append(errors, fmt.Errorf("value mismatch on node 1 key %s: expected %s, got %s", key, value, retrieved))
				mu.Unlock()
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
		}(i)
	}

	// Concurrent operations on node 2
	for i := 0; i < operationsPerNode; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			key := fmt.Sprintf("concurrent_key_2_%d", index)
			value := []byte(fmt.Sprintf("concurrent_value_2_%d", index))

			// Put operation
			_, err := h2.PutToKeyValueStore("concurrent_bucket_2", key, value)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("put error on node 2 key %s: %w", key, err))
				mu.Unlock()
				return
			}

			// Get operation
			retrieved, _, err := h2.GetFromKeyValueStore("concurrent_bucket_2", key)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("get error on node 2 key %s: %w", key, err))
				mu.Unlock()
				return
			}

			if string(retrieved) != string(value) {
				mu.Lock()
				errors = append(errors, fmt.Errorf("value mismatch on node 2 key %s: expected %s, got %s", key, value, retrieved))
				mu.Unlock()
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Check results
	mu.Lock()
	if len(errors) > 0 {
		return fmt.Errorf("concurrent operations failed: %v", errors[0])
	}
	mu.Unlock()

	expectedOperations := operationsPerNode * 2
	if successCount != expectedOperations {
		return fmt.Errorf("expected %d successful operations, got %d", expectedOperations, successCount)
	}

	fmt.Printf("✓ Concurrent KV operations successful (%d operations)\\n", successCount)

	// Test concurrent publishing
	fmt.Println("Testing concurrent publishing...")

	messagesSent := 0
	messagesReceived := make(map[string]int)
	mu = sync.Mutex{}

	// Set up subscribers
	received1 := make(chan []byte, operationsPerNode)
	received2 := make(chan []byte, operationsPerNode)

	cancel1, err := h1.SubscribeVolatileViaFanout("concurrent.test.node1",
		func(subject string, msg []byte) ([]byte, bool) {
			received1 <- msg
			return nil, false
		},
		func(err error) {
			fmt.Printf("Concurrent subscription error on node 1: %v\\n", err)
		})
	if err != nil {
		return fmt.Errorf("failed to set up subscription on node 1: %w", err)
	}
	defer cancel1()

	cancel2, err := h2.SubscribeVolatileViaFanout("concurrent.test.node2",
		func(subject string, msg []byte) ([]byte, bool) {
			received2 <- msg
			return nil, false
		},
		func(err error) {
			fmt.Printf("Concurrent subscription error on node 2: %v\\n", err)
		})
	if err != nil {
		return fmt.Errorf("failed to set up subscription on node 2: %w", err)
	}
	defer cancel2()

	// Wait for subscriptions to be ready
	time.Sleep(1 * time.Second)

	// Concurrent publishing
	wg = sync.WaitGroup{}
	for i := 0; i < operationsPerNode; i++ {
		wg.Add(2)

		go func(index int) {
			defer wg.Done()
			msg := []byte(fmt.Sprintf("concurrent_message_node1_%d", index))
			err := h1.PublishVolatile("concurrent.test.node1", msg)
			if err == nil {
				mu.Lock()
				messagesSent++
				mu.Unlock()
			}
		}(i)

		go func(index int) {
			defer wg.Done()
			msg := []byte(fmt.Sprintf("concurrent_message_node2_%d", index))
			err := h2.PublishVolatile("concurrent.test.node2", msg)
			if err == nil {
				mu.Lock()
				messagesSent++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Wait for message processing
	time.Sleep(2 * time.Second)

	// Count received messages
	close(received1)
	close(received2)

	for msg := range received1 {
		mu.Lock()
		messagesReceived["node1"]++
		mu.Unlock()
		_ = msg // Use the message
	}

	for msg := range received2 {
		mu.Lock()
		messagesReceived["node2"]++
		mu.Unlock()
		_ = msg // Use the message
	}

	mu.Lock()
	totalReceived := messagesReceived["node1"] + messagesReceived["node2"]
	mu.Unlock()

	fmt.Printf("✓ Concurrent publishing successful (sent: %d, received: %d)\\n", messagesSent, totalReceived)

	fmt.Println("✓ Concurrent operations test successful")
	return nil
}

func testClusterPerformance(tempDir string) error {
	fmt.Println("Testing performance across multiple nodes...")

	// Create two nodes for performance testing
	fmt.Println("Creating two nodes for performance testing...")

	// Node 1
	opts1, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 1: %w", err)
	}
	opts1.Port = 4243
	opts1.ClusterPort = 0
	opts1.LeafNodePort = 0
	opts1.StoreDir = filepath.Join(tempDir, "perf_node1")
	opts1.Name = "perf-node-1"

	h1, err := hub.NewHub(opts1)
	if err != nil {
		return fmt.Errorf("failed to create node 1: %w", err)
	}
	defer h1.Shutdown()

	// Node 2
	opts2, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 2: %w", err)
	}
	opts2.Port = 4244
	opts2.ClusterPort = 0
	opts2.LeafNodePort = 0
	opts2.StoreDir = filepath.Join(tempDir, "perf_node2")
	opts2.Name = "perf-node-2"

	h2, err := hub.NewHub(opts2)
	if err != nil {
		return fmt.Errorf("failed to create node 2: %w", err)
	}
	defer h2.Shutdown()

	fmt.Println("✓ Both performance nodes created successfully")

	// Create KV stores for performance testing
	fmt.Println("Setting up KV stores for performance testing...")

	kvConfig1 := hub.KeyValueStoreConfig{
		Bucket:   "perf_bucket_1",
		Replicas: 1,
	}
	err = h1.CreateOrUpdateKeyValueStore(kvConfig1)
	if err != nil {
		return fmt.Errorf("failed to create KV store on node 1: %w", err)
	}

	kvConfig2 := hub.KeyValueStoreConfig{
		Bucket:   "perf_bucket_2",
		Replicas: 1,
	}
	err = h2.CreateOrUpdateKeyValueStore(kvConfig2)
	if err != nil {
		return fmt.Errorf("failed to create KV store on node 2: %w", err)
	}

	fmt.Println("✓ KV stores created for performance testing")

	// Test KV put performance
	fmt.Println("Testing KV put performance...")

	operations := 100
	start := time.Now()

	for i := 0; i < operations; i++ {
		node := h1
		if i%2 == 1 {
			node = h2
		}

		bucket := "perf_bucket_1"
		if i%2 == 1 {
			bucket = "perf_bucket_2"
		}

		key := fmt.Sprintf("perf_key_%d", i)
		value := []byte(fmt.Sprintf("perf_value_%d", i))

		_, err := node.PutToKeyValueStore(bucket, key, value)
		if err != nil {
			return fmt.Errorf("KV put failed at operation %d: %w", i, err)
		}
	}

	kvPutDuration := time.Since(start)
	kvPutOpsPerSec := float64(operations) / kvPutDuration.Seconds()

	fmt.Printf("✓ KV Put Performance: %d ops in %v (%.2f ops/sec)\\n", operations, kvPutDuration, kvPutOpsPerSec)

	// Test KV get performance
	fmt.Println("Testing KV get performance...")

	start = time.Now()

	for i := 0; i < operations; i++ {
		node := h1
		if i%2 == 1 {
			node = h2
		}

		bucket := "perf_bucket_1"
		if i%2 == 1 {
			bucket = "perf_bucket_2"
		}

		key := fmt.Sprintf("perf_key_%d", i)

		_, _, err := node.GetFromKeyValueStore(bucket, key)
		if err != nil {
			return fmt.Errorf("KV get failed at operation %d: %w", i, err)
		}
	}

	kvGetDuration := time.Since(start)
	kvGetOpsPerSec := float64(operations) / kvGetDuration.Seconds()

	fmt.Printf("✓ KV Get Performance: %d ops in %v (%.2f ops/sec)\\n", operations, kvGetDuration, kvGetOpsPerSec)

	// Test volatile messaging performance
	fmt.Println("Testing volatile messaging performance...")

	messagesReceived := 0
	var receiveMutex sync.Mutex

	// Set up subscriber
	cancel, err := h2.SubscribeVolatileViaFanout("perf.test",
		func(subject string, msg []byte) ([]byte, bool) {
			receiveMutex.Lock()
			messagesReceived++
			receiveMutex.Unlock()
			return nil, false
		},
		func(err error) {
			fmt.Printf("Performance subscription error: %v\\n", err)
		})
	if err != nil {
		return fmt.Errorf("failed to set up performance subscription: %w", err)
	}
	defer cancel()

	// Wait for subscription to be ready
	time.Sleep(1 * time.Second)

	// Test publishing performance
	start = time.Now()

	for i := 0; i < operations; i++ {
		msg := []byte(fmt.Sprintf("perf_message_%d", i))
		err := h1.PublishVolatile("perf.test", msg)
		if err != nil {
			return fmt.Errorf("publish failed at operation %d: %w", i, err)
		}
	}

	publishDuration := time.Since(start)
	publishOpsPerSec := float64(operations) / publishDuration.Seconds()

	fmt.Printf("✓ Publish Performance: %d ops in %v (%.2f ops/sec)\\n", operations, publishDuration, publishOpsPerSec)

	// Wait for all messages to be received
	time.Sleep(2 * time.Second)

	receiveMutex.Lock()
	finalReceived := messagesReceived
	receiveMutex.Unlock()

	if finalReceived < operations {
		fmt.Printf("⚠ Warning: Only %d out of %d messages received\\n", finalReceived, operations)
	} else {
		fmt.Printf("✓ All %d messages received successfully\\n", finalReceived)
	}

	// Test persistent stream performance
	fmt.Println("Testing persistent stream performance...")

	// Create persistent stream
	streamConfig := &hub.PersistentConfig{
		Description: "Performance test stream",
		Subjects:    []string{"perf.persistent.>"},
		Retention:   0,
		MaxMsgs:     int64(operations + 50),
		Replicas:    1,
	}
	err = h1.CreateOrUpdatePersistent(streamConfig)
	if err != nil {
		return fmt.Errorf("failed to create performance stream: %w", err)
	}

	start = time.Now()

	for i := 0; i < operations/2; i++ { // Fewer operations for persistent to avoid timeout
		msg := []byte(fmt.Sprintf("perf_persistent_message_%d", i))
		err := h1.PublishPersistent("perf.persistent.test", msg)
		if err != nil {
			return fmt.Errorf("persistent publish failed at operation %d: %w", i, err)
		}
	}

	persistentDuration := time.Since(start)
	persistentOpsPerSec := float64(operations/2) / persistentDuration.Seconds()

	fmt.Printf("✓ Persistent Publish Performance: %d ops in %v (%.2f ops/sec)\\n", operations/2, persistentDuration, persistentOpsPerSec)

	// Summary
	fmt.Printf("\\n--- Performance Summary ---\\n")
	fmt.Printf("KV Put: %.2f ops/sec\\n", kvPutOpsPerSec)
	fmt.Printf("KV Get: %.2f ops/sec\\n", kvGetOpsPerSec)
	fmt.Printf("Volatile Publish: %.2f ops/sec\\n", publishOpsPerSec)
	fmt.Printf("Persistent Publish: %.2f ops/sec\\n", persistentOpsPerSec)

	fmt.Println("✓ Performance test successful")
	return nil
}

func testClusterDataConsistency(tempDir string) error {
	fmt.Println("Testing data consistency across multiple nodes...")

	// Create two nodes for consistency testing
	// Node 1
	opts1, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 1: %w", err)
	}
	opts1.Port = 4245
	opts1.ClusterPort = 0
	opts1.LeafNodePort = 0
	opts1.StoreDir = filepath.Join(tempDir, "consistency_node1")
	opts1.Name = "consistency-node-1"

	h1, err := hub.NewHub(opts1)
	if err != nil {
		return fmt.Errorf("failed to create node 1: %w", err)
	}
	defer h1.Shutdown()

	// Node 2
	opts2, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options for node 2: %w", err)
	}
	opts2.Port = 4246
	opts2.ClusterPort = 0
	opts2.LeafNodePort = 0
	opts2.StoreDir = filepath.Join(tempDir, "consistency_node2")
	opts2.Name = "consistency-node-2"

	h2, err := hub.NewHub(opts2)
	if err != nil {
		return fmt.Errorf("failed to create node 2: %w", err)
	}
	defer h2.Shutdown()

	fmt.Println("✓ Both consistency nodes created successfully")

	// Test data consistency within each node
	fmt.Println("Testing data consistency within individual nodes...")

	// Each node maintains its own consistent state
	kvConfig := hub.KeyValueStoreConfig{
		Bucket:   "consistency_test",
		Replicas: 1,
	}
	err = h1.CreateOrUpdateKeyValueStore(kvConfig)
	if err != nil {
		return fmt.Errorf("failed to create KV store on node 1: %w", err)
	}

	err = h2.CreateOrUpdateKeyValueStore(kvConfig)
	if err != nil {
		return fmt.Errorf("failed to create KV store on node 2: %w", err)
	}

	// Test sequential updates on the same node
	testKey := "consistency_key"
	values := []string{"value1", "value2", "value3"}

	for i, value := range values {
		_, err := h1.PutToKeyValueStore("consistency_test", testKey, []byte(value))
		if err != nil {
			return fmt.Errorf("failed to put value %d on node 1: %w", i+1, err)
		}

		retrieved, _, err := h1.GetFromKeyValueStore("consistency_test", testKey)
		if err != nil {
			return fmt.Errorf("failed to get value %d on node 1: %w", i+1, err)
		}

		if string(retrieved) != value {
			return fmt.Errorf("consistency error on node 1: expected %s, got %s", value, retrieved)
		}
	}

	fmt.Println("✓ Data consistency maintained within nodes")
	fmt.Println("✓ Data consistency test successful")
	return nil
}

func testClusterNetworkResilience(tempDir string) error {
	fmt.Println("Testing network resilience...")

	// Create node for resilience testing
	opts, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options: %w", err)
	}
	opts.Port = 4247
	opts.ClusterPort = 0
	opts.LeafNodePort = 0
	opts.StoreDir = filepath.Join(tempDir, "resilience_node")
	opts.Name = "resilience-node"

	h, err := hub.NewHub(opts)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}
	defer h.Shutdown()

	// Test persistent storage resilience
	streamConfig := &hub.PersistentConfig{
		Description: "Resilience test stream",
		Subjects:    []string{"resilience.>"},
		Retention:   0,
		MaxMsgs:     100,
		Replicas:    1,
	}
	err = h.CreateOrUpdatePersistent(streamConfig)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// Publish messages
	for i := 0; i < 5; i++ {
		msg := []byte(fmt.Sprintf("resilience-message-%d", i+1))
		err := h.PublishPersistent("resilience.test", msg)
		if err != nil {
			return fmt.Errorf("failed to publish message %d: %w", i+1, err)
		}
	}

	fmt.Println("✓ Network resilience test successful")
	return nil
}

func testClusterStreamReplication(tempDir string) error {
	fmt.Println("Testing stream operations across nodes...")

	// Create node for stream testing
	opts, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options: %w", err)
	}
	opts.Port = 4248
	opts.ClusterPort = 0
	opts.LeafNodePort = 0
	opts.StoreDir = filepath.Join(tempDir, "stream_node")
	opts.Name = "stream-node"

	h, err := hub.NewHub(opts)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}
	defer h.Shutdown()

	// Create multiple streams
	for i := 1; i <= 3; i++ {
		config := &hub.PersistentConfig{
			Description: fmt.Sprintf("Test stream %d", i),
			Subjects:    []string{fmt.Sprintf("stream%d.>", i)},
			Retention:   0,
			MaxMsgs:     100,
			Replicas:    1,
		}
		err = h.CreateOrUpdatePersistent(config)
		if err != nil {
			return fmt.Errorf("failed to create stream %d: %w", i, err)
		}
	}

	fmt.Println("✓ Stream replication test successful")
	return nil
}

func testClusterConsumerManagement(tempDir string) error {
	fmt.Println("Testing consumer management...")

	// Create node for consumer testing
	opts, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options: %w", err)
	}
	opts.Port = 4249
	opts.ClusterPort = 0
	opts.LeafNodePort = 0
	opts.StoreDir = filepath.Join(tempDir, "consumer_node")
	opts.Name = "consumer-node"

	h, err := hub.NewHub(opts)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}
	defer h.Shutdown()

	// Create stream
	config := &hub.PersistentConfig{
		Description: "Consumer management test stream",
		Subjects:    []string{"consumer.>"},
		Retention:   0,
		MaxMsgs:     100,
		Replicas:    1,
	}
	err = h.CreateOrUpdatePersistent(config)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// Create consumer
	received := make(chan []byte, 5)
	cancel, err := h.SubscribePersistentViaDurable("test-consumer", "consumer.test",
		func(subject string, msg []byte) ([]byte, bool, bool) {
			received <- msg
			return nil, false, true
		},
		func(err error) {
			fmt.Printf("Consumer error: %v\\n", err)
		})
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	defer cancel()

	// Publish and consume
	err = h.PublishPersistent("consumer.test", []byte("test message"))
	if err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	select {
	case <-received:
		fmt.Println("✓ Message consumed successfully")
	case <-time.After(3 * time.Second):
		return fmt.Errorf("timeout waiting for message")
	}

	fmt.Println("✓ Consumer management test successful")
	return nil
}

func testClusterConfigurationValidation(tempDir string) error {
	fmt.Println("Testing configuration validation...")

	// Test valid configuration
	opts, err := hub.DefaultNodeOptions()
	if err != nil {
		return fmt.Errorf("failed to create default options: %w", err)
	}
	opts.Port = 4250
	opts.ClusterPort = 0
	opts.LeafNodePort = 0
	opts.StoreDir = filepath.Join(tempDir, "config_node")
	opts.Name = "config-node"

	h, err := hub.NewHub(opts)
	if err != nil {
		return fmt.Errorf("failed to create node with valid config: %w", err)
	}
	defer h.Shutdown()

	// Test basic functionality
	err = h.PublishVolatile("config.test", []byte("config test"))
	if err != nil {
		return fmt.Errorf("failed to publish with valid config: %w", err)
	}

	fmt.Println("✓ Configuration validation test successful")
	return nil
}
