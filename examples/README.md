# Flin Examples

This directory contains example deployments for Flin.

## 10-Node Cluster

A production-scale distributed Flin cluster with 10 nodes.

**Quick Start:**
```bash
cd 10-node-cluster
./start.sh
```

**Features:**
- 10 nodes with Raft consensus
- 128 partitions for data distribution
- 3x replication for fault tolerance
- Can survive 7 node failures
- Expected: ~300K+ ops/sec cluster-wide

**Test it:**
```bash
cd ../..
./test-10-node-cluster.sh
```

See [10-node-cluster/README.md](./10-node-cluster/README.md) for details.
