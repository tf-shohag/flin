package db

import "errors"

// Common errors
var (
	ErrDocumentNotFound  = errors.New("document not found")
	ErrInvalidQuery      = errors.New("invalid query")
	ErrInvalidCollection = errors.New("invalid collection")
	ErrInvalidDocument   = errors.New("invalid document")
)

// Document represents a single document in a collection
type Document map[string]interface{}

// Query represents a filter condition
type Query struct {
	Field    string
	Operator string // "eq", "ne", "gt", "gte", "lt", "lte", "in"
	Value    interface{}
}

// FindOptions represents options for find operations
type FindOptions struct {
	Filters   []Query
	Sort      *SortOption
	Skip      int
	Limit     int
	IndexName string // Which index to use for optimization
}

// SortOption represents sorting configuration
type SortOption struct {
	Field     string
	Direction string // "asc" or "desc"
}

// UpdateOptions represents update operations
type UpdateOptions struct {
	Set   Document // Fields to set
	Unset []string // Fields to remove
	Merge bool     // If true, merge with existing doc; if false, replace
}

// IndexMetadata represents index configuration
type IndexMetadata struct {
	Collection string
	Field      string
	CreatedAt  int64
}

// Operator constants
const (
	OpEq  = "eq"
	OpNe  = "ne"
	OpGt  = "gt"
	OpGte = "gte"
	OpLt  = "lt"
	OpLte = "lte"
	OpIn  = "in"
)

// Sort direction constants
const (
	SortAsc  = "asc"
	SortDesc = "desc"
)
