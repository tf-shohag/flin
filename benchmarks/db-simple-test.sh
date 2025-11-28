#!/bin/bash

echo "🧪 Simple Document Store Test"
echo "=============================="
echo ""

# Save current directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR/.."

# Kill any existing flin-server processes
pkill -f flin-server 2>/dev/null || true
sleep 1

# Start the server in background
echo "🔧 Starting Flin server..."
./bin/flin-server \
  -node-id=test-node \
  -http=localhost:7080 \
  -raft=localhost:7090 \
  -port=:7380 \
  -data=./data/test \
  -partitions=4 \
  -workers=16 &

SERVER_PID=$!
echo "   Server PID: $SERVER_PID"

# Wait for server to start
echo "⏳ Waiting for server to start..."
sleep 5

# Check if server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "❌ Server failed to start"
    exit 1
fi

echo "✅ Server is running"
echo ""

# Create test program
mkdir -p /tmp/flin_doc_test
cd /tmp/flin_doc_test

cat > go.mod <<MODEOF
module flin-doc-test

go 1.21

require github.com/skshohagmiah/flin v0.0.0

replace github.com/skshohagmiah/flin => $SCRIPT_DIR/..
MODEOF

cat > main.go <<'EOF'
package main

import (
	"fmt"
	"log"

	flin "github.com/skshohagmiah/flin/clients/go"
)

func main() {
	// Create client
	client, err := flin.NewDocClient("localhost:7380")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	fmt.Println("📝 Testing Document Store Operations")
	fmt.Println("=====================================")
	fmt.Println()

	// Test 1: Insert a document
	fmt.Println("🟡 Test 1: INSERT")
	doc := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}
	id, err := client.Collection("users").Create().Data(doc).Exec()
	if err != nil {
		fmt.Printf("   ❌ INSERT failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ INSERT successful: id=%s\n", id)
	}
	fmt.Println()

	// Test 2: Find documents
	fmt.Println("🟢 Test 2: FIND")
	results, err := client.Collection("users").FindMany().
		Where("name", flin.Eq, "John Doe").
		Take(10).
		Exec()
	if err != nil {
		fmt.Printf("   ❌ FIND failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ FIND successful: found %d documents\n", len(results))
		for i, doc := range results {
			fmt.Printf("      [%d] %v\n", i, doc)
		}
	}
	fmt.Println()

	// Test 3: Update document
	fmt.Println("🔵 Test 3: UPDATE")
	err = client.Collection("users").Update().
		Where("name", flin.Eq, "John Doe").
		Set("age", 31).
		Exec()
	if err != nil {
		fmt.Printf("   ❌ UPDATE failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ UPDATE successful\n")
	}
	fmt.Println()

	// Test 4: Delete document
	fmt.Println("🔴 Test 4: DELETE")
	err = client.Collection("users").Delete().
		Where("name", flin.Eq, "John Doe").
		Exec()
	if err != nil {
		fmt.Printf("   ❌ DELETE failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ DELETE successful\n")
	}
	fmt.Println()

	fmt.Println("✅ All tests completed!")
}
EOF

echo "📊 Running simple document test..."
echo ""

go mod tidy 2>/dev/null
go run main.go

# Cleanup
echo ""
echo "🧹 Cleaning up..."
cd "$SCRIPT_DIR/.."
kill $SERVER_PID 2>/dev/null
rm -rf /tmp/flin_doc_test
rm -rf ./data/test

echo "✅ Test complete!"
