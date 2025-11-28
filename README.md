# 🚀 Flin - High-Performance Distributed Data Platform

A blazing-fast, distributed data platform combining **Key-Value Store**, **Message Queue**, **Stream Processing**, and **Document Database** in a single unified system.

## ⚡ Performance Highlights

| Component | Throughput | Latency | Notes |
|-----------|------------|---------|-------|
| **KV Store** | 319K reads/sec | 3.1μs | 3x faster than Redis |
| **KV Store** | 151K writes/sec | 6.6μs | Disk-backed durability |
| **Message Queue** | 104K push/sec | 9.6μs | Unified port with KV |
| **Message Queue** | 100K pop/sec | 10μs | BadgerDB persistence |
| **Stream** | High throughput | Low latency | Kafka-like pub/sub |
| **Document DB** | 76K inserts/sec | 13μs | MongoDB-like API |


## 🎯 Key Features

### 🔑 Key-Value Store
- ✅ **319K ops/sec** read throughput
- ✅ **151K ops/sec** write throughput  
- ✅ **Sub-10μs latency**
- ✅ **Atomic batch operations** (MSET/MGET/MDEL)
- ✅ **Dual storage**: Disk (durable) or Memory (fastest)
- ✅ **Text + Binary protocols** with auto-detection
- ✅ **Distributed clustering** with Raft consensus

### 📬 Message Queue
- ✅ **104K ops/sec** push throughput
- ✅ **100K ops/sec** pop throughput
- ✅ **Unified Port**: Runs on same port as KV (7380)
- ✅ **Durable**: Backed by BadgerDB
- ✅ **Atomic**: Crash-safe metadata management
- ✅ **Multiple queues** with independent operations

### 🌊 Stream Processing
- ✅ **Kafka-like** pub/sub messaging
- ✅ **Partitioned topics** for scalability
- ✅ **Consumer groups** with automatic rebalancing
- ✅ **Offset management** for reliable delivery
- ✅ **Retention policies** for automatic cleanup
- ✅ **At-least-once** delivery semantics

### 📄 Document Database
- ✅ **76K inserts/sec** throughput
- ✅ **13μs average latency**
- ✅ **MongoDB-like** document model
- ✅ **Prisma-like** fluent query builder
- ✅ **Secondary indexes** for fast queries
- ✅ **Flexible schema** with JSON documents
- ✅ **ACID transactions** via BadgerDB

## 🏗️ Architecture

Flin uses a **modular, layered architecture**:

```
┌─────────────────────────────────────────────────┐
│           Client SDKs (Go, Python, etc)         │
└─────────────────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────┐
│        Binary Protocol (Auto-detection)         │
└─────────────────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────┐
│  Server Layer (Hybrid: Fast Path + Workers)     │
│  ├─ KV Handlers                                 │
│  ├─ Queue Handlers                              │
│  ├─ Stream Handlers                             │
│  └─ Document Handlers                           │
└─────────────────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────┐
│         High-Level Abstraction Layer            │
│  ├─ internal/kv      (KV operations)            │
│  ├─ internal/queue   (Queue operations)         │
│  ├─ internal/stream  (Stream operations)        │
│  └─ internal/db      (Document operations)      │
└─────────────────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────┐
│            Storage Layer (BadgerDB)             │
│  ├─ internal/storage/kv.go                      │
│  ├─ internal/storage/queue.go                   │
│  ├─ internal/storage/stream.go                  │
│  └─ internal/storage/db.go                      │
└─────────────────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────┐
│        ClusterKit (Raft Consensus)              │
│  ├─ Leader Election                             │
│  ├─ Partition Management                        │
│  └─ Replication                                 │
└─────────────────────────────────────────────────┘
```

## 📦 Quick Start

### 🐳 Docker (Recommended)

**Single Node:**
```bash
cd docker/single && ./run.sh
```

**3-Node Cluster:**
```bash
cd docker/cluster && ./run.sh
```

Both scripts automatically:
- Start the node(s)
- Run performance benchmarks
- Show throughput metrics
- Leave cluster running for testing

See [docker/README.md](docker/README.md) for details.

### 💻 Local Installation

```bash
git clone https://github.com/skshohagmiah/flin
cd flin
go build -o bin/flin-server ./cmd/server
```

### Run Server

```bash
# Single node with all features
./bin/flin-server \
  -node-id=node1 \
  -http=:8080 \
  -raft=:9080 \
  -port=:7380 \
  -data=./data/node1 \
  -workers=256

# Join existing cluster
./bin/flin-server \
  -node-id=node2 \
  -http=:8081 \
  -raft=:9081 \
  -port=:7381 \
  -data=./data/node2 \
  -join=localhost:8080
```

## 💻 Unified Client Usage

Flin provides a single, unified client for all operations.

```go
import flin "github.com/skshohagmiah/flin/clients/go"

// Create unified client (connects to port 7380)
opts := flin.DefaultOptions("localhost:7380")
client, _ := flin.NewClient(opts)
defer client.Close()

// ============ 🔑 KV Store ============
client.KV.Set("user:1", []byte("John Doe"))
value, _ := client.KV.Get("user:1")
client.KV.Delete("user:1")

// Batch operations
client.KV.MSet([]string{"k1", "k2"}, [][]byte{[]byte("v1"), []byte("v2")})
values, _ := client.KV.MGet([]string{"k1", "k2"})

// ============ 📬 Message Queue ============
client.Queue.Push("tasks", []byte("Task 1"))
client.Queue.Push("tasks", []byte("Task 2"))

msg, _ := client.Queue.Pop("tasks")
fmt.Printf("Received: %s\n", string(msg))

// ============ 🌊 Stream Processing ============
// Create topic with 4 partitions and 7 days retention
client.Stream.CreateTopic("events", 4, 7*24*60*60*1000)

// Publish messages
client.Stream.Publish("events", -1, "user123", []byte(`{"action":"login"}`))

// Subscribe consumer group
client.Stream.Subscribe("events", "processors", "worker-1")

// Consume messages
messages, _ := client.Stream.Consume("events", "processors", "worker-1", 10)
for _, msg := range messages {
    fmt.Printf("Partition %d, Offset %d: %s\n", msg.Partition, msg.Offset, msg.Value)
    // Commit offset after processing
    client.Stream.Commit("events", "processors", msg.Partition, msg.Offset+1)
}

// ============ 📄 Document Database ============
// Insert document
id, _ := client.DB.Insert("users", map[string]interface{}{
    "name":  "John Doe",
    "email": "john@example.com",
    "age":   30,
})

// Find documents (Prisma-like API)
users, _ := client.DB.Query("users").
    Where("age", flin.Gte, 18).
    Where("status", flin.Eq, "active").
    OrderBy("created_at", flin.Desc).
    Skip(0).
    Take(10).
    Exec()

// Update document
client.DB.Update("users").
    Where("email", flin.Eq, "john@example.com").
    Set("age", 31).
    Set("verified", true).
    Exec()

// Delete document
client.DB.Delete("users").
    Where("status", flin.Eq, "inactive").
    Exec()
```

## 🔥 Performance Benchmarks

### KV Store
```bash
cd benchmarks
./kv-throughput.sh
```

**Results:**
- Read: 319K ops/sec (3.1μs latency)
- Write: 151K ops/sec (6.6μs latency)
- Batch: 792K ops/sec (1.26μs latency)

### Message Queue
```bash
./queue-throughput.sh
```

**Results:**
- Push: 104K ops/sec (9.6μs latency)
- Pop: 100K ops/sec (10μs latency)

### Stream Processing
```bash
./stream-throughput.sh
```

**Results:**
- High throughput pub/sub
- Efficient partition management
- Low-latency message delivery

### Document Database
```bash
./db-throughput.sh
```

**Results:**
- Insert: 76K docs/sec (13μs latency)
- Query: Fast with secondary indexes
- Update: Efficient in-place updates

## 📊 Performance vs Redis

| Operation | Flin | Redis | Speedup |
|-----------|------|-------|---------|
| KV Read | 319K/s | ~100K/s | **3.2x** |
| KV Write | 151K/s | ~80K/s | **1.9x** |
| Queue Push | 104K/s | ~80K/s | **1.3x** |
| Queue Pop | 100K/s | ~80K/s | **1.25x** |
| Batch Ops | 792K/s | ~100K/s | **7.9x** |

## 🛠️ Configuration

### Server Options

| Flag | Default | Description |
|------|---------|-------------|
| `-node-id` | (required) | Unique node identifier |
| `-http` | `:8080` | HTTP API address |
| `-raft` | `:9080` | Raft consensus address |
| `-port` | `:7380` | Unified server port (KV+Queue+Stream+Doc) |
| `-data` | `./data` | Data directory |
| `-workers` | `64` | Worker pool size |
| `-partitions` | `64` | Number of partitions |
| `-memory` | `false` | Use in-memory storage (no persistence) |
| `-join` | (empty) | Address of node to join |

### Storage Modes

**Disk Mode (Default):**
- Durable persistence via BadgerDB
- Survives restarts
- Optimized for throughput

**Memory Mode:**
- Fastest performance
- Data lost on restart
- Use for caching/temporary data

```bash
# Memory mode
./bin/flin-server -node-id=node1 -port=:7380 -memory
```

## 🏛️ Project Structure

```
flin/
├── cmd/
│   └── server/          # Server entry point
├── internal/
│   ├── kv/              # KV store abstraction
│   ├── queue/           # Queue abstraction
│   ├── stream/          # Stream abstraction
│   ├── db/              # Document store abstraction
│   │   ├── types.go     # Type definitions
│   │   ├── query.go     # Query builder
│   │   ├── helpers.go   # Utility functions
│   │   └── db.go        # Main implementation
│   ├── storage/         # Storage layer
│   │   ├── kv.go        # KV BadgerDB ops
│   │   ├── queue.go     # Queue BadgerDB ops
│   │   ├── stream.go    # Stream BadgerDB ops
│   │   └── db.go        # Document BadgerDB ops
│   ├── server/          # Server handlers
│   ├── protocol/        # Binary protocol
│   └── net/             # Connection pooling
├── clients/
│   └── go/              # Go client SDK
├── benchmarks/          # Performance tests
└── docker/              # Docker configs
```

## 🔐 Clustering & Replication

Flin uses **Raft consensus** for:
- Leader election
- Log replication
- Partition management
- Automatic failover

**3-Node Cluster Example:**
```bash
# Node 1 (bootstrap)
./bin/flin-server -node-id=node1 -http=:8080 -raft=:9080 -port=:7380

# Node 2 (join)
./bin/flin-server -node-id=node2 -http=:8081 -raft=:9081 -port=:7381 -join=localhost:8080

# Node 3 (join)
./bin/flin-server -node-id=node3 -http=:8082 -raft=:9082 -port=:7382 -join=localhost:8080
```

## 📚 Documentation

- [Architecture Overview](flow.md) - End-to-end data flow
- [Performance Summary](FINAL_PERFORMANCE_SUMMARY.md) - Detailed benchmarks
- [Docker Deployment](DOCKER.md) - Container setup
- [Benchmarks](benchmarks/) - Performance tests

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details

## 🙏 Acknowledgments

- Built with [BadgerDB](https://github.com/dgraph-io/badger) for storage
- Uses [ClusterKit](https://github.com/skshohagmiah/clusterkit) for Raft consensus
- Inspired by Redis, Kafka, and MongoDB

---

**Made with ❤️ by the Flin team**
