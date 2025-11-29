package storage

import (
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

// ShardedKVStorage distributes keys across multiple independent storage shards
// This eliminates global lock contention by routing keys to dedicated shards
// with independent BadgerDB instances
type ShardedKVStorage struct {
	shards []*Storage
	mu     []sync.RWMutex // Per-shard locks
	count  int
}

// NewShardedKVStorage creates a new sharded KV store with N independent shards
// Each shard manages keys independently, enabling true parallelism
func NewShardedKVStorage(path string, shardCount int) (*ShardedKVStorage, error) {
	if shardCount <= 0 || shardCount > 256 {
		return nil, fmt.Errorf("invalid shard count: %d (must be 1-256)", shardCount)
	}

	shards := make([]*Storage, shardCount)
	locks := make([]sync.RWMutex, shardCount)

	// Create independent BadgerDB instance for each shard
	for i := 0; i < shardCount; i++ {
		shardPath := fmt.Sprintf("%s/shard_%d", path, i)
		store, err := NewKVStorage(shardPath)
		if err != nil {
			// Clean up previously created shards
			for j := 0; j < i; j++ {
				shards[j].Close()
			}
			return nil, fmt.Errorf("failed to create shard %d: %w", i, err)
		}
		shards[i] = store
	}

	return &ShardedKVStorage{
		shards: shards,
		mu:     locks,
		count:  shardCount,
	}, nil
}

// getShard returns the shard for a given key using hash-based routing
// Uses FNV hash which is fast and provides good distribution
func (s *ShardedKVStorage) getShard(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32()) % s.count
}

// Set stores a key-value pair in the appropriate shard
func (s *ShardedKVStorage) Set(key string, value []byte, ttl time.Duration) error {
	shardID := s.getShard(key)
	s.mu[shardID].Lock()
	defer s.mu[shardID].Unlock()
	return s.shards[shardID].Set(key, value, ttl)
}

// Get retrieves a value by key from the appropriate shard
func (s *ShardedKVStorage) Get(key string) ([]byte, error) {
	shardID := s.getShard(key)
	s.mu[shardID].RLock()
	defer s.mu[shardID].RUnlock()
	return s.shards[shardID].Get(key)
}

// Delete removes a key from the appropriate shard
func (s *ShardedKVStorage) Delete(key string) error {
	shardID := s.getShard(key)
	s.mu[shardID].Lock()
	defer s.mu[shardID].Unlock()
	return s.shards[shardID].Delete(key)
}

// Exists checks if a key exists in the appropriate shard
func (s *ShardedKVStorage) Exists(key string) (bool, error) {
	shardID := s.getShard(key)
	s.mu[shardID].RLock()
	defer s.mu[shardID].RUnlock()
	return s.shards[shardID].Exists(key)
}

// Incr increments a numeric value in the appropriate shard
func (s *ShardedKVStorage) Incr(key string) error {
	shardID := s.getShard(key)
	s.mu[shardID].Lock()
	defer s.mu[shardID].Unlock()
	return s.shards[shardID].Incr(key)
}

// Decr decrements a numeric value in the appropriate shard
func (s *ShardedKVStorage) Decr(key string) error {
	shardID := s.getShard(key)
	s.mu[shardID].Lock()
	defer s.mu[shardID].Unlock()
	return s.shards[shardID].Decr(key)
}

// Scan retrieves all values with keys matching the given prefix from all shards
// Note: Requires scanning all shards, may be expensive for large datasets
func (s *ShardedKVStorage) Scan(prefix string) ([][]byte, error) {
	var allValues [][]byte
	
	// Scan all shards in parallel for better performance
	errChan := make(chan error, s.count)
	valuesChan := make(chan [][]byte, s.count)

	for i := 0; i < s.count; i++ {
		go func(shardID int) {
			s.mu[shardID].RLock()
			defer s.mu[shardID].RUnlock()
			values, err := s.shards[shardID].Scan(prefix)
			if err != nil {
				errChan <- err
			} else {
				valuesChan <- values
			}
		}(i)
	}

	// Collect results
	for i := 0; i < s.count; i++ {
		select {
		case err := <-errChan:
			return nil, err
		case values := <-valuesChan:
			allValues = append(allValues, values...)
		}
	}

	return allValues, nil
}

// ScanKeys retrieves all keys matching the given prefix from all shards
func (s *ShardedKVStorage) ScanKeys(prefix string) ([]string, error) {
	var allKeys []string

	// Scan all shards in parallel
	errChan := make(chan error, s.count)
	keysChan := make(chan []string, s.count)

	for i := 0; i < s.count; i++ {
		go func(shardID int) {
			s.mu[shardID].RLock()
			defer s.mu[shardID].RUnlock()
			keys, err := s.shards[shardID].ScanKeys(prefix)
			if err != nil {
				errChan <- err
			} else {
				keysChan <- keys
			}
		}(i)
	}

	// Collect results
	for i := 0; i < s.count; i++ {
		select {
		case err := <-errChan:
			return nil, err
		case keys := <-keysChan:
			allKeys = append(allKeys, keys...)
		}
	}

	return allKeys, nil
}

// ScanKeysWithValues retrieves all keys and values matching the given prefix from all shards
func (s *ShardedKVStorage) ScanKeysWithValues(prefix string) (map[string][]byte, error) {
	allKVs := make(map[string][]byte)
	mu := sync.Mutex{}

	// Scan all shards in parallel
	errChan := make(chan error, s.count)
	kvsChan := make(chan map[string][]byte, s.count)

	for i := 0; i < s.count; i++ {
		go func(shardID int) {
			s.mu[shardID].RLock()
			defer s.mu[shardID].RUnlock()
			kvs, err := s.shards[shardID].ScanKeysWithValues(prefix)
			if err != nil {
				errChan <- err
			} else {
				kvsChan <- kvs
			}
		}(i)
	}

	// Collect results
	for i := 0; i < s.count; i++ {
		select {
		case err := <-errChan:
			return nil, err
		case kvs := <-kvsChan:
			mu.Lock()
			for k, v := range kvs {
				allKVs[k] = v
			}
			mu.Unlock()
		}
	}

	return allKVs, nil
}

// BatchSet stores multiple key-value pairs, distributed across shards
// Groups keys by shard to minimize cross-shard overhead
func (s *ShardedKVStorage) BatchSet(kvPairs map[string][]byte, ttl time.Duration) error {
	// Group keys by shard
	shardedPairs := make([]map[string][]byte, s.count)
	for i := 0; i < s.count; i++ {
		shardedPairs[i] = make(map[string][]byte)
	}

	for key, value := range kvPairs {
		shardID := s.getShard(key)
		shardedPairs[shardID][key] = value
	}

	// Set batches in parallel across shards
	errChan := make(chan error, s.count)
	for i := 0; i < s.count; i++ {
		go func(shardID int) {
			if len(shardedPairs[shardID]) == 0 {
				errChan <- nil
				return
			}
			s.mu[shardID].Lock()
			defer s.mu[shardID].Unlock()
			errChan <- s.shards[shardID].BatchSet(shardedPairs[shardID], ttl)
		}(i)
	}

	// Collect errors
	for i := 0; i < s.count; i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}
	return nil
}

// BatchGet retrieves multiple values, distributed across shards
// Groups keys by shard to minimize cross-shard overhead
func (s *ShardedKVStorage) BatchGet(keys []string) (map[string][]byte, error) {
	// Group keys by shard
	shardedKeys := make([][]string, s.count)
	for i := 0; i < s.count; i++ {
		shardedKeys[i] = make([]string, 0)
	}

	for _, key := range keys {
		shardID := s.getShard(key)
		shardedKeys[shardID] = append(shardedKeys[shardID], key)
	}

	// Get values in parallel across shards
	results := make(chan map[string][]byte, s.count)
	errChan := make(chan error, s.count)

	for i := 0; i < s.count; i++ {
		go func(shardID int) {
			if len(shardedKeys[shardID]) == 0 {
				results <- make(map[string][]byte)
				errChan <- nil
				return
			}
			s.mu[shardID].RLock()
			defer s.mu[shardID].RUnlock()
			kvs, err := s.shards[shardID].BatchGet(shardedKeys[shardID])
			results <- kvs
			errChan <- err
		}(i)
	}

	// Collect results
	allResults := make(map[string][]byte)
	for i := 0; i < s.count; i++ {
		if err := <-errChan; err != nil {
			return nil, err
		}
		for k, v := range <-results {
			allResults[k] = v
		}
	}

	return allResults, nil
}

// BatchDelete removes multiple keys, distributed across shards
func (s *ShardedKVStorage) BatchDelete(keys []string) error {
	// Group keys by shard
	shardedKeys := make([][]string, s.count)
	for i := 0; i < s.count; i++ {
		shardedKeys[i] = make([]string, 0)
	}

	for _, key := range keys {
		shardID := s.getShard(key)
		shardedKeys[shardID] = append(shardedKeys[shardID], key)
	}

	// Delete batches in parallel across shards
	errChan := make(chan error, s.count)
	for i := 0; i < s.count; i++ {
		go func(shardID int) {
			if len(shardedKeys[shardID]) == 0 {
				errChan <- nil
				return
			}
			s.mu[shardID].Lock()
			defer s.mu[shardID].Unlock()
			errChan <- s.shards[shardID].BatchDelete(shardedKeys[shardID])
		}(i)
	}

	// Collect errors
	for i := 0; i < s.count; i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}
	return nil
}

// Close closes all shard instances
func (s *ShardedKVStorage) Close() error {
	var lastErr error
	for i := 0; i < s.count; i++ {
		if err := s.shards[i].Close(); err != nil {
			lastErr = err // Keep trying to close all shards
		}
	}
	return lastErr
}
