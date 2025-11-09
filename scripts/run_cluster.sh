#!/bin/bash

# Run a 3-node Flin distributed cluster locally

set -e

cd "$(dirname "$0")/.."

echo "🚀 Starting Flin Distributed Cluster (3 nodes)"
echo "==============================================="
echo ""

# Clean up old data
echo "🧹 Cleaning up old data..."
rm -rf /tmp/flin-node-*

# Build the kvserver binary
echo "📦 Building kvserver..."
go build -o /tmp/kvserver ./cmd/kvserver

echo ""
echo "Starting nodes..."
echo ""

# Start node 1 (bootstrap)
echo "🔷 Node 1 (bootstrap) - HTTP:8080 Raft:9080 KV:6380"
/tmp/kvserver \
  -node-id=node-1 \
  -http=:8080 \
  -raft=:9080 \
  -port=:6380 \
  -data=/tmp/flin-node-1 \
  > /tmp/flin-node-1.log 2>&1 &
NODE1_PID=$!

sleep 3

# Start node 2 (join)
echo "🔷 Node 2 (joining)   - HTTP:8081 Raft:9081 KV:6381"
/tmp/kvserver \
  -node-id=node-2 \
  -http=:8081 \
  -raft=:9081 \
  -port=:6381 \
  -join=localhost:8080 \
  -data=/tmp/flin-node-2 \
  > /tmp/flin-node-2.log 2>&1 &
NODE2_PID=$!

sleep 3

# Start node 3 (join)
echo "🔷 Node 3 (joining)   - HTTP:8082 Raft:9082 KV:6382"
/tmp/kvserver \
  -node-id=node-3 \
  -http=:8082 \
  -raft=:9082 \
  -port=:6382 \
  -join=localhost:8080 \
  -data=/tmp/flin-node-3 \
  > /tmp/flin-node-3.log 2>&1 &
NODE3_PID=$!

sleep 3

echo ""
echo "✅ Cluster started!"
echo ""
echo "📊 Cluster Info:"
echo "   Node 1: Cluster:8080 KV:6380 (PID: $NODE1_PID)"
echo "   Node 2: Cluster:8081 KV:6381 (PID: $NODE2_PID)"
echo "   Node 3: Cluster:8082 KV:6382 (PID: $NODE3_PID)"
echo ""
echo "📚 Try these commands:"
echo ""
echo "   # Check cluster topology"
echo "   curl http://localhost:8080/cluster | jq"
echo ""
echo "   # Check health"
echo "   curl http://localhost:8080/health | jq"
echo ""
echo "   # View logs"
echo "   tail -f /tmp/flin-node-1.log"
echo ""
echo "Press Ctrl+C to stop all nodes..."
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping cluster..."
    kill $NODE1_PID $NODE2_PID $NODE3_PID 2>/dev/null || true
    sleep 2
    echo "✅ Cluster stopped"
    exit 0
}

trap cleanup INT TERM

# Wait for all processes
wait
