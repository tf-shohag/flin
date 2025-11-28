package main

import (
	"fmt"
	"log"

	flin "github.com/skshohagmiah/flin/clients/go"
)

func main() {
	// Create unified client
	opts := flin.DefaultOptions("localhost:7380")
	client, err := flin.NewClient(opts)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	fmt.Println("🚀 Testing Unified Client API")
	fmt.Println("============================")

	// 1. KV Store
	fmt.Println("\n🔑 Testing KV Store...")
	err = client.KV.Set("test:key", []byte("Hello Flin"))
	if err != nil {
		log.Fatalf("KV Set failed: %v", err)
	}
	val, err := client.KV.Get("test:key")
	if err != nil {
		log.Fatalf("KV Get failed: %v", err)
	}
	fmt.Printf("   ✅ Set/Get: %s\n", string(val))

	// 2. Queue
	fmt.Println("\n📬 Testing Queue...")
	err = client.Queue.Push("test:queue", []byte("Task 1"))
	if err != nil {
		log.Fatalf("Queue Push failed: %v", err)
	}
	msg, err := client.Queue.Pop("test:queue")
	if err != nil {
		log.Fatalf("Queue Pop failed: %v", err)
	}
	fmt.Printf("   ✅ Push/Pop: %s\n", string(msg))

	// 3. Stream
	fmt.Println("\n🌊 Testing Stream...")
	err = client.Stream.CreateTopic("test:stream", 1, 3600000)
	if err != nil {
		fmt.Printf("   ⚠️  CreateTopic: %v (might already exist)\n", err)
	}
	err = client.Stream.Publish("test:stream", 0, "k1", []byte("Event 1"))
	if err != nil {
		log.Fatalf("Stream Publish failed: %v", err)
	}
	fmt.Printf("   ✅ Publish successful\n")

	// 4. Document DB
	fmt.Println("\n📄 Testing Document DB...")
	id, err := client.DB.Insert("users", map[string]interface{}{
		"name": "Alice",
		"age":  25,
	})
	if err != nil {
		log.Fatalf("DB Insert failed: %v", err)
	}
	fmt.Printf("   ✅ Inserted ID: %s\n", id)

	results, err := client.DB.Query("users").
		Where("name", flin.Eq, "Alice").
		Exec()
	if err != nil {
		log.Fatalf("DB Query failed: %v", err)
	}
	fmt.Printf("   ✅ Query found %d docs\n", len(results))

	fmt.Println("\n✨ All tests passed!")
}
