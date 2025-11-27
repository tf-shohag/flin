package main

import (
	"fmt"
	"log"
	"time"

	flin "github.com/skshohagmiah/flin/clients/go"
)

func main() {
	// Create client
	opts := flin.DefaultOptions("localhost:7380")
	client, err := flin.NewClient(opts)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if client.Queue == nil {
		log.Fatal("Queue client is nil!")
	}

	fmt.Println("✓ Client created successfully")
	fmt.Printf("✓ Queue client: %v\n", client.Queue != nil)

	// Test PUSH
	testData := []byte("Hello, Queue!")
	fmt.Println("\n📤 Testing PUSH...")
	err = client.Queue.Push("test_queue", testData)
	if err != nil {
		log.Fatalf("PUSH failed: %v", err)
	}
	fmt.Println("✓ PUSH successful")

	// Test POP
	fmt.Println("\n📥 Testing POP...")
	value, err := client.Queue.Pop("test_queue")
	if err != nil {
		log.Fatalf("POP failed: %v", err)
	}
	fmt.Printf("✓ POP successful: %s\n", string(value))

	// Test multiple operations
	fmt.Println("\n🔄 Testing 10 PUSH operations...")
	start := time.Now()
	for i := 0; i < 10; i++ {
		err = client.Queue.Push("bench_queue", []byte(fmt.Sprintf("msg-%d", i)))
		if err != nil {
			log.Fatalf("PUSH %d failed: %v", i, err)
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("✓ 10 PUSH operations completed in %v\n", elapsed)

	fmt.Println("\n✅ All tests passed!")
}
