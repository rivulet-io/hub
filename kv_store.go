package hub

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type KeyValueStoreConfig struct {
	Bucket       string
	Description  string
	MaxValueSize Size
	TTL          time.Duration
	MaxBytes     Size
	Replicas     int
}

func (h *Hub) CreateOrUpdateKeyValueStore(config KeyValueStoreConfig) error {
	storeConfig := &nats.KeyValueConfig{
		Bucket:       config.Bucket,
		Description:  config.Description,
		MaxValueSize: int32(config.MaxValueSize.Bytes()),
		TTL:          config.TTL,
		MaxBytes:     config.MaxBytes.Bytes(),
		Replicas:     config.Replicas,
		Storage:      nats.FileStorage,
		Placement: &nats.Placement{
			Cluster: HubClusterName,
		},
		History: 1,
	}
	_, err := h.jetstreamCtx.CreateKeyValue(storeConfig)
	if err != nil {
		return fmt.Errorf("failed to create or update key-value store: %w", err)
	}

	return nil
}

func (h *Hub) GetFromKeyValueStore(bucket, key string) ([]byte, uint64, error) {
	kv, err := h.jetstreamCtx.KeyValue(bucket)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to access key-value store %q: %w", bucket, err)
	}

	entry, err := kv.Get(key)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get key %q from bucket %q: %w", key, bucket, err)
	}

	return entry.Value(), entry.Revision(), nil
}

func (h *Hub) PutToKeyValueStore(bucket, key string, value []byte) (uint64, error) {
	kv, err := h.jetstreamCtx.KeyValue(bucket)
	if err != nil {
		return 0, fmt.Errorf("failed to access key-value store %q: %w", bucket, err)
	}

	revision, err := kv.Put(key, value)
	if err != nil {
		return 0, fmt.Errorf("failed to put key %q to bucket %q: %w", key, bucket, err)
	}

	return revision, nil
}

func (h *Hub) UpdateToKeyValueStore(bucket, key string, value []byte, expectedRevision uint64) (uint64, error) {
	kv, err := h.jetstreamCtx.KeyValue(bucket)
	if err != nil {
		return 0, fmt.Errorf("failed to access key-value store %q: %w", bucket, err)
	}

	revision, err := kv.Update(key, value, expectedRevision)
	if err != nil {
		return 0, fmt.Errorf("failed to update key %q in bucket %q: %w", key, bucket, err)
	}

	return revision, nil
}

func (h *Hub) DeleteFromKeyValueStore(bucket, key string) error {
	kv, err := h.jetstreamCtx.KeyValue(bucket)
	if err != nil {
		return fmt.Errorf("failed to access key-value store %q: %w", bucket, err)
	}

	if err := kv.Delete(key); err != nil {
		return fmt.Errorf("failed to delete key %q from bucket %q: %w", key, bucket, err)
	}

	return nil
}

func (h *Hub) PurgeKeyValueStore(bucket, key string) error {
	kv, err := h.jetstreamCtx.KeyValue(bucket)
	if err != nil {
		return fmt.Errorf("failed to access key-value store %q: %w", bucket, err)
	}

	if err := kv.Purge(key); err != nil {
		return fmt.Errorf("failed to purge key %q from bucket %q: %w", key, bucket, err)
	}

	return nil
}
