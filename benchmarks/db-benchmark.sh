#!/bin/bash

# Document Store Benchmark Script
# Tests database performance with various workloads

set -e

echo "🧪 Flin Document Store Benchmarks"
echo "=================================="
echo ""

# Get the project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Start the server if not already running
echo "📦 Starting Flin server..."
if ! pgrep -f "flin.*server" > /dev/null; then
    ./bin/flin-server -node-id=benchmark-node -http=:8090 -raft=:9090 -port=:6390 &
    SERVER_PID=$!
    sleep 2
    echo "✅ Server started (PID: $SERVER_PID)"
    CLEANUP=true
else
    echo "ℹ️  Server already running"
    CLEANUP=false
fi

# Cleanup function
cleanup() {
    if [ "$CLEANUP" = true ]; then
        echo ""
        echo "🛑 Stopping server..."
        kill $SERVER_PID || true
    fi
}

trap cleanup EXIT

echo ""
echo "📊 Running benchmarks..."
echo ""

# Run Go benchmarks
echo "Running Go unit benchmarks..."
go test -bench=. -benchmem ./internal/db/... -v -run=^$ 2>&1 | tee /tmp/db-benchmark.log

echo ""
echo "✅ Benchmarks complete!"
echo ""
echo "Summary:"
echo "--------"
grep "Benchmark" /tmp/db-benchmark.log

echo ""
echo "📁 Full results saved to: /tmp/db-benchmark.log"
