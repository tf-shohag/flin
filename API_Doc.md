# 📚 Flin Client API Documentation

This document provides a detailed reference for the Flin Go Client API. The client is unified, meaning a single client instance gives you access to Key-Value, Queue, Stream, and Document Store operations.

## 🚀 Initialization

### Creating a Client

```go
import flin "github.com/skshohagmiah/flin/clients/go"

// Connect to a single node or unified port
opts := flin.DefaultOptions("localhost:7380")

// Optional: Configure connection pool
opts.MinConnections = 16
opts.MaxConnections = 256

client, err := flin.NewClient(opts)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

---

## 🔑 Key-Value Store (`client.KV`)

High-performance key-value storage with support for strings, binary data, and atomic counters.

### `Set(key string, value []byte) error`
Stores a value for a given key. Overwrites existing values.
```go
err := client.KV.Set("user:101", []byte("Alice"))
```

### `Get(key string) ([]byte, error)`
Retrieves the value for a key. Returns error if key does not exist.
```go
val, err := client.KV.Get("user:101")
```

### `Delete(key string) error`
Removes a key and its value.
```go
err := client.KV.Delete("user:101")
```

### `Exists(key string) (bool, error)`
Checks if a key exists.
```go
exists, err := client.KV.Exists("user:101")
```

### `Incr(key string) (int64, error)`
Atomically increments a counter. Creates the key if it doesn't exist.
```go
newVal, err := client.KV.Incr("visits:page:home")
```

### `Decr(key string) (int64, error)`
Atomically decrements a counter.
```go
newVal, err := client.KV.Decr("stock:item:123")
```

### Batch Operations
- **`MSet(keys []string, values [][]byte) error`**: Set multiple keys atomically.
- **`MGet(keys []string) ([][]byte, error)`**: Get multiple keys at once.
- **`MDelete(keys []string) error`**: Delete multiple keys at once.

---

## 📬 Message Queue (`client.Queue`)

Persistent, ordered message queue for task processing and decoupling services.

### `Push(queue string, item []byte) error`
Adds an item to the end of the queue.
```go
err := client.Queue.Push("email_tasks", []byte(`{"to":"user@example.com"}`))
```

### `Pop(queue string) ([]byte, error)`
Removes and returns the item from the front of the queue. Blocks if empty (optional implementation dependent).
```go
task, err := client.Queue.Pop("email_tasks")
```

### `Peek(queue string) ([]byte, error)`
Returns the item at the front without removing it.
```go
task, err := client.Queue.Peek("email_tasks")
```

### `Len(queue string) (int64, error)`
Returns the number of items in the queue.
```go
count, err := client.Queue.Len("email_tasks")
```

### `Clear(queue string) error`
Removes all items from the queue.
```go
err := client.Queue.Clear("email_tasks")
```

---

## 🌊 Stream Processing (`client.Stream`)

Kafka-like pub/sub system with topics, partitions, and consumer groups.

### `CreateTopic(topic string, partitions int, retentionMs int64) error`
Creates a new topic with specified partitions and retention period.
```go
// 4 partitions, 7 days retention
err := client.Stream.CreateTopic("logs", 4, 7*24*60*60*1000)
```

### `Publish(topic string, partition int, key string, value []byte) error`
Publishes a message to a topic. Use `partition: -1` for automatic partitioning based on key hash.
```go
err := client.Stream.Publish("logs", -1, "server-1", []byte("Error: 500"))
```

### `Subscribe(topic, group, consumer string) error`
Registers a consumer as part of a consumer group.
```go
err := client.Stream.Subscribe("logs", "log-processors", "worker-1")
```

### `Consume(topic, group, consumer string, count int) ([]StreamMessage, error)`
Fetches a batch of messages for the consumer.
```go
msgs, err := client.Stream.Consume("logs", "log-processors", "worker-1", 10)
for _, msg := range msgs {
    fmt.Printf("Partition: %d, Offset: %d, Value: %s\n", msg.Partition, msg.Offset, msg.Value)
}
```

### `Commit(topic, group string, partition int, offset uint64) error`
Commits the processed offset for a consumer group.
```go
err := client.Stream.Commit("logs", "log-processors", msg.Partition, msg.Offset+1)
```

---

## 📄 Document Database (`client.DB`)

MongoDB-like document store with a Prisma-like fluent query builder.

### `Insert(collection string, doc map[string]interface{}) (string, error)`
Inserts a JSON document and returns its generated ID (UUID).
```go
id, err := client.DB.Insert("users", map[string]interface{}{
    "name": "John Doe",
    "age":  30,
    "active": true,
})
```

### `Query(collection string)`
Starts a query builder chain.

#### Methods:
- **`Where(field, operator, value)`**: Add a filter condition.
- **`OrderBy(field, direction)`**: Sort results.
- **`Skip(n)`**: Skip first n results.
- **`Take(n)`**: Limit results to n.
- **`Exec()`**: Execute the query.

#### Operators:
`flin.Eq`, `flin.Ne`, `flin.Gt`, `flin.Gte`, `flin.Lt`, `flin.Lte`, `flin.In`

#### Example:
```go
users, err := client.DB.Query("users").
    Where("age", flin.Gte, 18).
    Where("active", flin.Eq, true).
    OrderBy("created_at", flin.Desc).
    Take(20).
    Exec()
```

### `Update(collection string)`
Starts an update builder chain.

#### Methods:
- **`Where(...)`**: Select documents to update.
- **`Set(field, value)`**: Set field values.
- **`Exec()`**: Execute the update.

#### Example:
```go
err := client.DB.Update("users").
    Where("id", flin.Eq, "user-123").
    Set("active", false).
    Set("updated_at", time.Now()).
    Exec()
```

### `Delete(collection string)`
Starts a delete builder chain.

#### Methods:
- **`Where(...)`**: Select documents to delete.
- **`Exec()`**: Execute the delete.

#### Example:
```go
err := client.DB.Delete("users").
    Where("active", flin.Eq, false).
    Exec()
```

---

## ⚠️ Error Handling

All methods return standard Go `error` types. Common errors include:
- `key not found`: The requested key does not exist.
- `connection refused`: Server is unreachable.
- `timeout`: Operation took too long.

Always check errors!

```go
if err != nil {
    log.Printf("Operation failed: %v", err)
    return
}
```
