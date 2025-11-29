package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/skshohagmiah/flin/internal/kv"
)

func main() {
	fmt.Println("🔬 KV Sharding Performance Test")
	fmt.Println("================================")
	fmt.Println()

	concurrency := 64
	duration := 5 * time.Second
	valueSize := 1024

	// Test 1: Non-sharded (baseline)
	fmt.Println("📊 Test 1: Non-Sharded Storage")
	fmt.Println("------------------------------")
	testNonSharded(concurrency, duration, valueSize)
	fmt.Println()

	// Test 2: Sharded with 64 shards
	fmt.Println("📊 Test 2: Sharded Storage (64 shards)")
	fmt.Println("--------------------------------------")
	testSharded(concurrency, duration, valueSize, 64)
	fmt.Println()

	// Test 3: Sharded with 16 shards (lower overhead)
	fmt.Println("📊 Test 3: Sharded Storage (16 shards)")
	fmt.Println("--------------------------------------")
	testSharded(concurrency, duration, valueSize, 16)
}

func testNonSharded(concurrency int, duration time.Duration, valueSize int) {
	store, err := kv.New("./data/kv_test_nonsharded")
	if err != nil {
		fmt.Printf("❌ Failed to create store: %v\n", err)
		return
	}
	defer store.Close()

	value := make([]byte, valueSize)
	for i := range value {
		value[i] = byte(i % 256)
	}

	// WRITE test
	fmt.Println("🔴 WRITE Test")
	var writeOps atomic.Int64
	var wg sync.WaitGroup

	startTime := time.Now()
	stopTime := startTime.Add(duration)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ops := int64(0)
			for time.Now().Before(stopTime) {
				key := fmt.Sprintf("key_%d_%d", workerID, ops)
				if err := store.Set(key, value, 0); err == nil {
					ops++
				}
			}
			writeOps.Add(ops)
		}(i)
	}

	wg.Wait()
	writeElapsed := time.Since(startTime)
	writeTotal := writeOps.Load()
	writeThroughput := float64(writeTotal) / writeElapsed.Seconds()

	fmt.Printf("Operations: %.2fK\n", float64(writeTotal)/1000)
	fmt.Printf("Throughput: %.2fK ops/sec\n", writeThroughput/1000)
	fmt.Printf("Latency:    %.2fμs\n", (writeElapsed.Seconds()*1000000)/float64(writeTotal))
}

func testSharded(concurrency int, duration time.Duration, valueSize int, shardCount int) {
	store, err := kv.NewSharded(fmt.Sprintf("./data/kv_test_sharded_%d", shardCount), shardCount)
	if err != nil {
		fmt.Printf("❌ Failed to create sharded store: %v\n", err)
		return
	}
	defer store.Close()

	value := make([]byte, valueSize)
	for i := range value {
		value[i] = byte(i % 256)
	}

	// WRITE test
	fmt.Println("🔴 WRITE Test")
	var writeOps atomic.Int64
	var wg sync.WaitGroup

	startTime := time.Now()
	stopTime := startTime.Add(duration)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ops := int64(0)
			for time.Now().Before(stopTime) {
				key := fmt.Sprintf("key_%d_%d", workerID, ops)
				if err := store.Set(key, value, 0); err == nil {
					ops++
				}
			}
			writeOps.Add(ops)
		}(i)
	}

	wg.Wait()
	writeElapsed := time.Since(startTime)
	writeTotal := writeOps.Load()
	writeThroughput := float64(writeTotal) / writeElapsed.Seconds()

	fmt.Printf("Operations: %.2fK\n", float64(writeTotal)/1000)
	fmt.Printf("Throughput: %.2fK ops/sec\n", writeThroughput/1000)
	fmt.Printf("Latency:    %.2fμs\n", (writeElapsed.Seconds()*1000000)/float64(writeTotal))
}
