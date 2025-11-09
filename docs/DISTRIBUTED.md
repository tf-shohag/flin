# 🌐 Flin - Distributed KV Store

**Flin is a distributed-first KV store** powered by ClusterKit for coordination.

## Quick Start

### Start a 3-Node Cluster

```bash
# Node 1 (bootstrap - first node)
./kvserver -node-id=node-1 -http=:8080 -raft=:9080 -port=:6380

# Node 2 (join cluster)
./kvserver -node-id=node-2 -http=:8081 -raft=:9081 -port=:6381 -join=localhost:8080

# Node 3 (join cluster)
./kvserver -node-id=node-3 -http=:8082 -raft=:9082 -port=:6382 -join=localhost:8080
```

**That's it!** You now have a fault-tolerant, distributed KV store.

## Client Usage

### Cluster-Aware Client

```go
import "github.com/skshohagmiah/flin/pkg/client"

// Connect to cluster (provide any node addresses)
c, err := client.NewClusterClient([]string{
    "localhost:8080",
    "localhost:8081",
    "localhost:8082",
})
defer c.Close()

// Client automatically:
// - Fetches cluster topology
// - Routes to correct partition owner
// - Handles failover
// - Replicates writes

c.Set("user:123", []byte("data"))
value, _ := c.Get("user:123")
```

### How It Works

1. **Topology Discovery** - Client fetches partition map from ClusterKit
2. **Partition Routing** - MD5(key) → partition → primary node
3. **Automatic Failover** - If primary fails, try replicas
4. **Background Replication** - Writes replicate async to replicas
5. **Topology Refresh** - Updates every 30s or on failure

## Architecture

```
Client Request
     ↓
Partition Calculation (MD5 hash)
     ↓
Primary Node (write) or Any Node (read)
     ↓
Local KV Store
     ↓
Async Replication to Replicas
```

## Features

✅ **Automatic Partitioning** - 64 partitions by default
✅ **Replication** - 3x replication for fault tolerance  
✅ **Raft Consensus** - ClusterKit handles coordination
✅ **Health Checking** - Automatic failure detection
✅ **Auto Rebalancing** - Partitions redistribute on topology changes
✅ **Topology-Aware Client** - Smart routing and failover

## Performance

- **Distributed**: 40-50K SET, 200-300K GET ops/sec per node
- **Latency**: 1-5ms (network + replication overhead)
- **Scalability**: Linear read scaling with more nodes

## Summary

**Flin is a distributed-first KV store** - built for fault tolerance and horizontal scaling from day one!
