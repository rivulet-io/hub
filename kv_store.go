package hub

import (
	"context"
	"errors"
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
		History:     1,
		Compression: true,
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

func (h *Hub) DeleteKeyValueStore(bucket string) error {
	if err := h.jetstreamCtx.DeleteKeyValue(bucket); err != nil {
		return fmt.Errorf("failed to delete key-value store %q: %w", bucket, err)
	}

	return nil
}

const lockValue = "locked"

func (h *Hub) TryLock(bucket, key string) (cancel func(), err error) {
	kv, err := h.jetstreamCtx.KeyValue(bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to access key-value store %q: %w", bucket, err)
	}

	revision, err := kv.Create(key, []byte(lockValue))
	if err != nil {
		return nil, fmt.Errorf("failed to lock key %q in bucket %q: %w", key, bucket, err)
	}

	return func() {
		_ = kv.Delete(key, nats.LastRevision(revision))
	}, nil
}

type LockOptions struct {
	initialDelay  time.Duration
	MaxDelay      time.Duration
	BackOffFactor int
}

func (h *Hub) Lock(ctx context.Context, bucket, key string, opt ...LockOptions) (cancel func(), err error) {
	option := LockOptions{
		initialDelay:  time.Millisecond * 10,
		MaxDelay:      2 * time.Second,
		BackOffFactor: 2,
	}
	if len(opt) > 0 {
		option = opt[0]
	}

	currentDelay := option.initialDelay
	backOffFactor := time.Duration(option.BackOffFactor)
	maxDelay := option.MaxDelay

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		cancel, err = h.TryLock(bucket, key)
		if err == nil {
			return cancel, nil
		}
		if !errors.Is(err, nats.ErrKeyExists) {
			return nil, fmt.Errorf("failed to lock key %q in bucket %q: %w", key, bucket, err)
		}
		time.Sleep(currentDelay)
		currentDelay *= backOffFactor
		if currentDelay > maxDelay {
			currentDelay = maxDelay
		}
	}
}

func (h *Hub) ForceUnlock(bucket, key string) error {
	kv, err := h.jetstreamCtx.KeyValue(bucket)
	if err != nil {
		return fmt.Errorf("failed to access key-value store %q: %w", bucket, err)
	}

	if err := kv.Delete(key); err != nil {
		return fmt.Errorf("failed to force unlock key %q in bucket %q: %w", key, bucket, err)
	}

	return nil
}

func (h *Hub) IsLocked(bucket, key string) (bool, error) {
	kv, err := h.jetstreamCtx.KeyValue(bucket)
	if err != nil {
		return false, fmt.Errorf("failed to access key-value store %q: %w", bucket, err)
	}

	_, err = kv.Get(key)
	if err != nil {
		if err == nats.ErrKeyNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to get key %q from bucket %q: %w", key, bucket, err)
	}

	return true, nil
}
