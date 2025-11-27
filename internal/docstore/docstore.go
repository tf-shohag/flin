package docstore

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
)

// DocStore manages document storage
type DocStore struct {
	db *badger.DB
	mu sync.RWMutex
}

// Document is a generic JSON object
type Document map[string]interface{}

// New creates a new DocStore
func New(path string) (*DocStore, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger: %w", err)
	}

	return &DocStore{db: db}, nil
}

// Close closes the store
func (d *DocStore) Close() error {
	return d.db.Close()
}

// Insert adds a document to a collection
func (d *DocStore) Insert(collection string, doc Document) (string, error) {
	// Generate ID if missing
	id, ok := doc["_id"].(string)
	if !ok || id == "" {
		id = uuid.New().String()
		doc["_id"] = id
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}

	err = d.db.Update(func(txn *badger.Txn) error {
		// Store document: doc:{collection}:{id}
		key := makeDocKey(collection, id)
		if err := txn.Set([]byte(key), data); err != nil {
			return err
		}

		// Update indexes (TODO: Load index definitions and update them)
		return nil
	})

	return id, err
}

// Get retrieves a document by ID
func (d *DocStore) Get(collection, id string) (Document, error) {
	var doc Document
	err := d.db.View(func(txn *badger.Txn) error {
		key := makeDocKey(collection, id)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &doc)
		})
	})
	return doc, err
}

// Find executes a query
// For MVP, we'll do a full collection scan if no index is used
func (d *DocStore) Find(collection string, query map[string]interface{}) ([]Document, error) {
	results := make([]Document, 0)

	err := d.db.View(func(txn *badger.Txn) error {
		prefix := []byte(fmt.Sprintf("doc:%s:", collection))
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var doc Document
				if err := json.Unmarshal(val, &doc); err != nil {
					return err
				}

				if matches(doc, query) {
					results = append(results, doc)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return results, err
}

// Update modifies documents matching a query
func (d *DocStore) Update(collection string, query map[string]interface{}, update map[string]interface{}) (int, error) {
	count := 0
	err := d.db.Update(func(txn *badger.Txn) error {
		// Scan and update (inefficient for large datasets, needs indexing)
		prefix := []byte(fmt.Sprintf("doc:%s:", collection))
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			// We need to copy key because item is only valid in this iteration
			key := item.KeyCopy(nil)

			err := item.Value(func(val []byte) error {
				var doc Document
				if err := json.Unmarshal(val, &doc); err != nil {
					return err
				}

				if matches(doc, query) {
					// Apply updates
					for k, v := range update {
						doc[k] = v
					}

					newData, err := json.Marshal(doc)
					if err != nil {
						return err
					}

					if err := txn.Set(key, newData); err != nil {
						return err
					}
					count++
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return count, err
}

// Delete removes documents matching a query
func (d *DocStore) Delete(collection string, query map[string]interface{}) (int, error) {
	count := 0
	err := d.db.Update(func(txn *badger.Txn) error {
		prefix := []byte(fmt.Sprintf("doc:%s:", collection))
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		keysToDelete := make([][]byte, 0)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var doc Document
				if err := json.Unmarshal(val, &doc); err != nil {
					return err
				}

				if matches(doc, query) {
					keysToDelete = append(keysToDelete, item.KeyCopy(nil))
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		for _, key := range keysToDelete {
			if err := txn.Delete(key); err != nil {
				return err
			}
			count++
		}
		return nil
	})
	return count, err
}

// Helpers

func makeDocKey(collection, id string) string {
	return fmt.Sprintf("doc:%s:%s", collection, id)
}

// matches checks if a document matches a simple query (exact match for now)
// Query format: {"field": "value", "age": {"$gt": 18}}
func matches(doc Document, query map[string]interface{}) bool {
	for k, v := range query {
		docVal, exists := doc[k]
		if !exists {
			return false
		}

		// Handle operators
		if queryMap, ok := v.(map[string]interface{}); ok {
			if !matchOperators(docVal, queryMap) {
				return false
			}
		} else {
			// Exact match
			if fmt.Sprintf("%v", docVal) != fmt.Sprintf("%v", v) {
				return false
			}
		}
	}
	return true
}

func matchOperators(val interface{}, ops map[string]interface{}) bool {
	for op, target := range ops {
		switch op {
		case "$gt":
			return compare(val, target) > 0
		case "$lt":
			return compare(val, target) < 0
		case "$gte":
			return compare(val, target) >= 0
		case "$lte":
			return compare(val, target) <= 0
		case "$ne":
			return fmt.Sprintf("%v", val) != fmt.Sprintf("%v", target)
		}
	}
	return true
}

// compare returns 1 if a > b, -1 if a < b, 0 if equal
// Very basic comparison for numbers and strings
func compare(a, b interface{}) int {
	// Try float64
	f1, ok1 := toFloat(a)
	f2, ok2 := toFloat(b)
	if ok1 && ok2 {
		if f1 > f2 {
			return 1
		}
		if f1 < f2 {
			return -1
		}
		return 0
	}

	// String comparison
	s1 := fmt.Sprintf("%v", a)
	s2 := fmt.Sprintf("%v", b)
	return strings.Compare(s1, s2)
}

func toFloat(v interface{}) (float64, bool) {
	switch i := v.(type) {
	case float64:
		return i, true
	case float32:
		return float64(i), true
	case int:
		return float64(i), true
	case int64:
		return float64(i), true
	case int32:
		return float64(i), true
	}
	return 0, false
}
