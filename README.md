# 🚀 Flin - High-Performance KV Store

Flin is a fast, distributed data engine built in Go, designed to handle key-value storage, message queues, and streaming workloads — all under one unified system.

**Current Status:** ✅ KV Store (Production Ready) | 🚧 Queue & Stream (Coming Soon)

## 🎯 Performance

**Optimized with NATS-style architecture:**
- ✅ **103K SET ops/sec** (4 workers)
- ✅ **787K GET ops/sec** (4 workers)
- ✅ **140K mixed ops/sec** (70% reads, 30% writes)
- ✅ **Sub-40μs latency** for writes
- ✅ **Sub-5μs latency** for reads

**Matches Redis performance for embedded use!** 🚀

See [BENCHMARKS.md](./BENCHMARKS.md) for detailed performance analysis.

## 🚀 Quick Start

### 1. Test Single Node

```bash
./test-single-node.sh
```

This will:
- Build and start a single Flin node
- Run a 10-second benchmark with 128 workers
- Show throughput and latency
- Clean up automatically

Expected: **~90K ops/sec**

### 2. Test 3-Node Cluster

```bash
./test-3-node-cluster.sh
```

This will:
- Start a 3-node Docker cluster with Raft consensus
- Wait for all nodes to be healthy
- Run benchmark on all nodes simultaneously (384 workers total)
- Show per-node and cluster-wide throughput
- Calculate scaling efficiency
- Clean up automatically

Expected: **~300K+ ops/sec** (cluster-wide, 3x scaling)

### 3. As an Embedded Library

```go
import "github.com/skshohagmiah/flin/internal/kv"
store, _ := kv.New("./data")
defer store.Close()

// Set with TTL
store.Set("session:123", []byte("data"), 3600*time.Second)

// Get
value, _ := store.Get("session:123")

// Batch operations (5-10x faster!)
batch := map[string][]byte{
    "key1": []byte("value1"),
    "key2": []byte("value2"),
}
store.BatchSet(batch)
```

### Using the Go SDK Client

```go
import "github.com/skshohagmiah/flin/pkg/client"

// Connect to Flin server
c, _ := client.New("localhost:6380")
defer c.Close()

// Simple operations
c.Set("mykey", []byte("hello"))
value, _ := c.Get("mykey")

// With connection pooling (recommended)
pc, _ := client.NewPooledClient(client.DefaultPoolConfig())
defer pc.Close()

pc.Set("key", []byte("value"))
```

See [pkg/client/README.md](./pkg/client/README.md) for full SDK documentation.

## 📦 Installation

### Docker

```bash
docker pull flin-kv:latest
docker run -d -p 6380:6380 -v flin-data:/data flin-kv:latest ./kvserver
```

### From Source

```bash
git clone https://github.com/skshohagmiah/flin.git
cd flin
make build
```

## 🧠 Architecture

Flin uses **NATS-style architecture** for extreme performance:

- **Inline processing** (no spawning overhead)
- **Buffered channels** for async I/O

## ⚡ Core Features

- 🌐 **Distributed** - Raft consensus with ClusterKit coordination
- 🔄 **Auto Partitioning** - 64 partitions with consistent hashing
- 🛡️ **Fault Tolerant** - 3x replication, survives node failures
- ⚖️ **Auto Rebalancing** - Partitions redistribute on topology changes
- 🚀 **High Performance** - 40-50K SET, 200-300K GET ops/sec per node
- 💾 **Persistent Storage** - BadgerDB LSM tree with caching
- 📡 **Topology-Aware Client** - Smart routing and automatic failover
- ⏱️ **TTL Support** - Automatic expiration
- 🐳 **Docker Ready** - Production-ready containers

| Operation | Throughput | Latency | Notes |
|-----------|------------|---------|-------|
| **SET** | 103K ops/sec | 38 μs | 4 workers optimal |
| **GET** | 787K ops/sec | 4 μs | Cache-optimized |
| **MIXED** | 140K ops/sec | 28 μs | 70% reads, 30% writes |
| **DELETE** | 165K ops/sec | 23 μs | Fast tombstones |
**Batch Operations:**
- BatchSet: 500K-1M+ ops/sec
- BatchGet: 1M-2M+ ops/sec

## 📚 Documentation

- [BENCHMARKS.md](./BENCHMARKS.md) - Detailed performance analysis
- [DOCKER.md](./DOCKER.md) - Docker deployment guide
- [performance.md](./performance.md) - Performance tuning guide
- [cmd/kvserver/README.md](./cmd/kvserver/README.md) - Server architecture

## 🛠️ Development

```bash
# Build everything
make build

# Run tests
make test

# Run benchmark
make benchmark

# Docker build
make docker-build

# Clean
make clean
```

## 🎯 Use Cases

### Session Store
```go
// 90% reads, 10% writes
// Throughput: ~600K ops/sec
// Latency: <5μs p99
store.Set("session:"+id, sessionData, 24*time.Hour)
```

### Cache Layer
```go
// 95% reads, 5% writes
// Throughput: ~700K ops/sec
// Latency: <4μs p99
store.Set("cache:"+key, data, 1*time.Hour)
```

### Event Store
```go
// 20% reads, 80% writes
// Throughput: ~90K ops/sec
// Latency: <50μs p99
store.Set("event:"+id, eventData, 0)
```

## 🔮 Roadmap

- ✅ **KV Store** - Production ready
- 🚧 **Queue** - Coming soon
- 🚧 **Stream** - Coming soon
- 🚧 **Clustering** - Planned
- 🚧 **Replication** - Planned

## 📄 License

MIT License - see LICENSE file for details

## 🤝 Contributing

Contributions welcome! Please open an issue or PR.

---

**Built with ❤️ using Go, BadgerDB, and NATS-style architecture**