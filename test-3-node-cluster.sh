#!/bin/bash

echo "🚀 Flin 3-Node Cluster Test"
echo "============================"
echo ""

# Save script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running"
    exit 1
fi

# Clean up any existing flin containers
echo "🧹 Cleaning up existing containers..."
docker ps -a --filter "name=flin-node" --format "{{.Names}}" | xargs -r docker rm -f 2>/dev/null || true
docker network rm 3-node-cluster_flin-cluster 10-node-cluster_flin-cluster 2>/dev/null || true

# Start the cluster
echo "🔧 Starting 3-node cluster..."
cd examples/3-node-cluster

# Clean up any existing cluster
docker compose down -v 2>/dev/null

# Start all 3 nodes
docker compose up -d

# Return to script directory
cd "$SCRIPT_DIR"

# Wait for nodes to be healthy
echo "⏳ Waiting for nodes to be healthy..."
for i in {1..60}; do
    healthy=$(docker ps --filter "name=flin-node" --filter "health=healthy" --format "{{.Names}}" | wc -l)
    echo -n "   Healthy nodes: $healthy/3"
    
    if [ "$healthy" -eq 3 ]; then
        echo " ✅"
        break
    fi
    
    echo ""
    sleep 2
done

# Check cluster info
echo ""
echo "📊 Cluster Info:"
curl -s http://localhost:8080/cluster | jq '{
  nodes: .cluster.nodes | length,
  partitions: .cluster.partition_map.partitions | length,
  replication: .cluster.config.replication_factor
}'

echo ""
echo "📊 Running benchmark..."
echo "   Concurrency: 128 workers per node (384 total)"
echo "   Duration: 10 seconds per node"
echo "   Value size: 1KB"
echo ""

# Create benchmark program
cat > /tmp/flin_cluster_bench.go << 'EOF'
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/skshohagmiah/flin/pkg/client"
)

func benchNode(port int, wg *sync.WaitGroup, results *[3]int64) {
	defer wg.Done()
	
	concurrency := 128
	duration := 10 * time.Second
	
	var totalOps atomic.Int64
	var workerWg sync.WaitGroup
	
	value := make([]byte, 1024)
	startTime := time.Now()
	stopTime := startTime.Add(duration)
	
	addr := fmt.Sprintf("localhost:%d", port)
	
	for i := 0; i < concurrency; i++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()
			
			c, err := client.NewTCP(addr)
			if err != nil {
				return
			}
			defer c.Close()
			
			ops := int64(0)
			for time.Now().Before(stopTime) {
				key := fmt.Sprintf("node%d_key_%d_%d", port, workerID, ops)
				if err := c.Set(key, value); err == nil {
					ops++
				}
			}
			totalOps.Add(ops)
		}(i)
	}
	
	workerWg.Wait()
	results[port-6380] = totalOps.Load()
}

func main() {
	var wg sync.WaitGroup
	var results [3]int64
	
	// Benchmark all 3 nodes in parallel
	for i := 0; i < 3; i++ {
		port := 6380 + i
		wg.Add(1)
		go benchNode(port, &wg, &results)
	}
	
	wg.Wait()
	
	// Calculate totals
	var total int64
	fmt.Println("\n📊 Results:")
	fmt.Println("===========\n")
	
	for i := 0; i < 3; i++ {
		port := 6380 + i
		ops := results[i]
		total += ops
		
		opsPerSec := float64(ops) / 10.0
		fmt.Printf("   Node %d (port %d): %8.2fK ops/sec\n", i+1, port, opsPerSec/1000)
	}
	
	fmt.Println("   ──────────────────────────────────")
	totalPerSec := float64(total) / 10.0
	fmt.Printf("   TOTAL CLUSTER:    %8.2fK ops/sec\n", totalPerSec/1000)
	
	fmt.Println("\n📈 Analysis:")
	fmt.Println("============\n")
	avgPerNode := totalPerSec / 3
	fmt.Printf("   Average per node: %8.2fK ops/sec\n", avgPerNode/1000)
	fmt.Printf("   Total cluster:    %8.2fK ops/sec\n", totalPerSec/1000)
	fmt.Printf("   Total workers:    384 (128 per node)\n")
	fmt.Printf("   Latency per op:   ~%.2fμs\n", (10.0*1000000)/totalPerSec)
	
	// Calculate scaling efficiency
	singleNodeThroughput := 135720.0 // From single node test
	scalingFactor := totalPerSec / singleNodeThroughput
	efficiency := (scalingFactor / 3.0) * 100
	
	fmt.Println("\n🎯 Scaling:")
	fmt.Println("===========\n")
	fmt.Printf("   Single node:      135.72K ops/sec\n")
	fmt.Printf("   3-node cluster:   %8.2fK ops/sec\n", totalPerSec/1000)
	fmt.Printf("   Scaling factor:   %.2fx (ideal: 3.0x)\n", scalingFactor)
	fmt.Printf("   Efficiency:       %.1f%%\n", efficiency)
}
EOF

go run /tmp/flin_cluster_bench.go

# Cleanup
echo ""
echo "🧹 Cleaning up..."
cd "$SCRIPT_DIR/examples/3-node-cluster"
docker compose down
cd "$SCRIPT_DIR"
rm -f /tmp/flin_cluster_bench.go

echo ""
echo "✅ Test complete!"
