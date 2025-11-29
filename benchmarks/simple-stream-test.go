package main

import (
"fmt"
"sync"
"sync/atomic"
"time"

flin "github.com/skshohagmiah/flin/clients/go"
)

func main() {
concurrency := 64
duration := 5 * time.Second
numPartitions := 64  // Use many partitions to allow true parallelism

opts := flin.DefaultOptions("localhost:7380")
opts.MaxConnections = concurrency + 10
opts.MinConnections = concurrency / 2

client, err := flin.NewClient(opts)
if err != nil {
fmt.Printf("Failed: %v\n", err)
return
}
defer client.Close()

// Create 1 topic with 64 partitions
fmt.Println("Creating topic with 64 partitions...")
client.Stream.CreateTopic("test_topic", numPartitions, 0)
time.Sleep(500 * time.Millisecond)

value := make([]byte, 1024)
for i := range value {
value[i] = byte(i % 256)
}

// Test PUBLISH
fmt.Println("🔴 PUBLISH Test")
var pubOps atomic.Int64
var wg sync.WaitGroup

startTime := time.Now()
stopTime := startTime.Add(duration)

for i := 0; i < concurrency; i++ {
wg.Add(1)
go func(workerID int) {
defer wg.Done()
partition := workerID % numPartitions  // Distribute across partitions
ops := int64(0)
for time.Now().Before(stopTime) {
if err := client.Stream.Publish("test_topic", partition, "", value); err == nil {
ops++
}
}
pubOps.Add(ops)
}(i)
}

wg.Wait()
pubElapsed := time.Since(startTime)
pubTotal := pubOps.Load()
pubThroughput := float64(pubTotal) / pubElapsed.Seconds()

fmt.Printf("Operations:  %.2fK\n", float64(pubTotal)/1000)
fmt.Printf("Throughput:  %.2fK ops/sec\n", pubThroughput/1000)
fmt.Printf("Latency:     %.2fμs\n", (pubElapsed.Seconds() * 1000000) / float64(pubTotal))
}
