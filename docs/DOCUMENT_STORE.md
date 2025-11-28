# Flin Document Store (DB)

The Document Store is a MongoDB-like document database built into Flin, providing schema-less JSON document storage with rich querying capabilities, secondary indexes, and a Prisma-inspired fluent client API.

## Features

- **JSON Documents**: Schema-less storage organized by collections
- **Auto-generated IDs**: UUID-based document identifiers
- **Secondary Indexes**: Index fields for optimized queries
- **Rich Queries**: Filtering, sorting, pagination
- **Fluent API**: Prisma-inspired chainable query builder
- **Durable Storage**: BadgerDB-backed persistence
- **Timestamps**: Automatic `_created_at` and `_updated_at` fields

## Architecture

### Storage Layer (`internal/db`)
- **DocStore**: Main API managing collections and documents
- **BadgerDB**: Underlying key-value store (optimized settings)
- **In-memory Indexes**: Fast field-based lookups

### Protocol (`pkg/protocol`)
Binary protocol opcodes (0x40-0x44):
- `OpDocInsert` (0x40): Insert a document
- `OpDocFind` (0x41): Query documents
- `OpDocUpdate` (0x42): Update documents
- `OpDocDelete` (0x43): Delete documents
- `OpDocIndex` (0x44): Create indexes (future)

### Client SDK (`clients/go`)
Fluent query builder with chainable operations for Create, Read, Update, Delete (CRUD).

## Storage Format

Documents are stored as JSON with special fields:
```json
{
  "_id": "550e8400-e29b-41d4-a716-446655440000",
  "_created_at": 1700000000,
  "_updated_at": 1700000000,
  "name": "John Doe",
  "age": 30,
  "email": "john@example.com"
}
```

**Storage Key Pattern**: `doc:<collection>:<id>`

## Database API

### Insert

```go
doc := db.Document{
    "name": "John Doe",
    "age": 30,
}
id, err := docStore.Insert("users", doc)
```

### Get

```go
doc, err := docStore.Get("users", id)
```

### Find (Query)

```go
opts := db.FindOptions{
    Filters: []db.Query{
        {Field: "age", Operator: "gt", Value: 25},
    },
    Sort: &db.SortOption{
        Field:     "name",
        Direction: "asc",
    },
    Skip:  0,
    Limit: 10,
}
results, err := docStore.Find("users", opts)
```

**Query Operators:**
- `"eq"` - Equal
- `"ne"` - Not equal
- `"gt"` - Greater than
- `"gte"` - Greater than or equal
- `"lt"` - Less than
- `"lte"` - Less than or equal
- `"in"` - In array

### FindOne

```go
doc, err := docStore.FindOne("users", opts)
```

### Update

```go
updateOpts := db.UpdateOptions{
    Set:   db.Document{"age": 31},
    Merge: true, // Merge with existing or replace
}
err := docStore.Update("users", id, updateOpts)
```

### Delete

```go
err := docStore.Delete("users", id)
```

### DeleteMany

```go
deleted, err := docStore.DeleteMany("users", opts)
```

### CreateIndex

```go
err := docStore.CreateIndex("users", "email")
```

Indexes are stored in memory and persisted via metadata. Manual index creation is required for optimal performance on non-ID fields.

## Go Client API

### Connect

```go
import "github.com/skshohagmiah/flin/clients/go"

client, err := NewDocClient("localhost:6380")
defer client.Close()
```

### Create

```go
id, err := client.Collection("users").Create().
    Set("name", "John Doe").
    Set("email", "john@example.com").
    Set("age", 30).
    Exec()
```

### Find Many

```go
users, err := client.Collection("users").FindMany().
    Where("age", Gt, 25).
    OrderBy("name", Asc).
    Skip(0).
    Take(10).
    Exec()
```

### Find Unique

```go
user, err := client.Collection("users").FindUnique().
    Where("email", Eq, "john@example.com").
    Exec()
```

### Update

```go
updated, err := client.Collection("users").Update().
    Where("age", Lt, 25).
    Set("status", "young").
    Exec()
```

### Delete

```go
deleted, err := client.Collection("users").Delete().
    Where("status", Eq, "inactive").
    Exec()
```

## Query Builder Operators

```go
const (
    Eq  = "eq"  // Equality
    Ne  = "ne"  // Not equal
    Gt  = "gt"  // Greater than
    Gte = "gte" // Greater than or equal
    Lt  = "lt"  // Less than
    Lte = "lte" // Less than or equal
    In  = "in"  // In array
)
```

## Sort Orders

```go
const (
    Asc  = "asc"  // Ascending order
    Desc = "desc" // Descending order
)
```

## Examples

### Example 1: User Management

```go
client, _ := NewDocClient("localhost:6380")

// Create users
for i := 0; i < 100; i++ {
    client.Collection("users").Create().
        Set("name", fmt.Sprintf("User %d", i)).
        Set("age", 20+rand.Intn(40)).
        Set("email", fmt.Sprintf("user%d@example.com", i)).
        Exec()
}

// Find active users over 25
users, _ := client.Collection("users").FindMany().
    Where("age", Gt, 25).
    Where("status", Eq, "active").
    OrderBy("name", Asc).
    Exec()

for _, user := range users {
    fmt.Printf("%v\n", user)
}
```

### Example 2: Pagination

```go
// Get page 2 with 20 items per page
page := 2
pageSize := 20

results, _ := client.Collection("products").FindMany().
    Skip((page - 1) * pageSize).
    Take(pageSize).
    OrderBy("created_at", Desc).
    Exec()
```

### Example 3: Bulk Operations

```go
// Delete inactive users
deleted, _ := client.Collection("users").Delete().
    Where("last_login", Lt, time.Now().AddDate(0, 0, -90)).
    Exec()

fmt.Printf("Deleted %d inactive users\n", deleted)
```

## Performance

Performance depends on collection size and query complexity:

### Benchmarks (1000 documents)

```
BenchmarkInsert-8    100000    12,456 ns/op
BenchmarkFind-8       50000    24,783 ns/op
BenchmarkUpdate-8     75000    15,234 ns/op
```

**Tips for optimization:**
1. Create indexes on frequently filtered fields
2. Use `Take()` to limit result sets
3. Use `Skip()` + `Take()` for pagination
4. Avoid complex filters on large collections

## Storage Location

By default, the document store is persisted at `./data/db` using BadgerDB.

Configure with server flag:
```bash
./bin/flin-server -data ./path/to/data
```

## Testing

Run tests:
```bash
go test -v ./internal/db/...
```

Run benchmarks:
```bash
go test -bench=. -benchmem ./internal/db/...
```

Or use the benchmark script:
```bash
./benchmarks/db-benchmark.sh
```

## Server Integration

The document store is automatically initialized when the Flin server starts. It's accessible via the binary protocol for remote clients.

### Server Startup

```go
// Automatically initialized in cmd/server/main.go
docStore, _ := db.New(dataDir + "/db")
srv, _ := server.NewServerWithWorkers(
    kvStore,
    queue,
    stream,
    docStore,  // Passed to server
    ck,
    port,
    nodeID,
    workers,
)
```

### Protocol Handling

Document operations are automatically dispatched in the server's binary request handler:

```go
case protocol.OpDocInsert:
    c.processBinaryDocInsert(req, startTime)
case protocol.OpDocFind:
    c.processBinaryDocFind(req, startTime)
case protocol.OpDocUpdate:
    c.processBinaryDocUpdate(req, startTime)
case protocol.OpDocDelete:
    c.processBinaryDocDelete(req, startTime)
```

## Future Enhancements

- [ ] Index management protocol operations (OpDocIndex)
- [ ] Aggregation pipeline support
- [ ] Transaction support (multi-document transactions)
- [ ] Full-text search indexing
- [ ] TTL (time-to-live) for automatic document expiration
- [ ] Schema validation
- [ ] Graph queries
- [ ] Distributed document store across cluster

## Related Documentation

- [BINARY_PROTOCOL.md](../docs/BINARY_PROTOCOL.md) - Protocol details
- [GETTING_STARTED.md](../docs/GETTING_STARTED.md) - General Flin setup
- [OPTIMAL_CONFIG.md](../docs/OPTIMAL_CONFIG.md) - Configuration tuning
