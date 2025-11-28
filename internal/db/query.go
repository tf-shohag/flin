package db

import (
	"fmt"
)

// QueryBuilder provides a fluent interface for building queries
type QueryBuilder struct {
	collection string
	filters    []Query
	sort       *SortOption
	skip       int
	limit      int
}

// NewQueryBuilder creates a new query builder for a collection
func NewQueryBuilder(collection string) *QueryBuilder {
	return &QueryBuilder{
		collection: collection,
		filters:    []Query{},
		limit:      -1,
	}
}

// Where adds a filter condition
func (qb *QueryBuilder) Where(field, operator string, value interface{}) *QueryBuilder {
	qb.filters = append(qb.filters, Query{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return qb
}

// WhereEq adds an equality filter (shorthand)
func (qb *QueryBuilder) WhereEq(field string, value interface{}) *QueryBuilder {
	return qb.Where(field, OpEq, value)
}

// WhereNe adds a not-equal filter (shorthand)
func (qb *QueryBuilder) WhereNe(field string, value interface{}) *QueryBuilder {
	return qb.Where(field, OpNe, value)
}

// WhereGt adds a greater-than filter (shorthand)
func (qb *QueryBuilder) WhereGt(field string, value interface{}) *QueryBuilder {
	return qb.Where(field, OpGt, value)
}

// WhereGte adds a greater-than-or-equal filter (shorthand)
func (qb *QueryBuilder) WhereGte(field string, value interface{}) *QueryBuilder {
	return qb.Where(field, OpGte, value)
}

// WhereLt adds a less-than filter (shorthand)
func (qb *QueryBuilder) WhereLt(field string, value interface{}) *QueryBuilder {
	return qb.Where(field, OpLt, value)
}

// WhereLte adds a less-than-or-equal filter (shorthand)
func (qb *QueryBuilder) WhereLte(field string, value interface{}) *QueryBuilder {
	return qb.Where(field, OpLte, value)
}

// WhereIn adds an in-array filter (shorthand)
func (qb *QueryBuilder) WhereIn(field string, values interface{}) *QueryBuilder {
	return qb.Where(field, OpIn, values)
}

// OrderBy sets the sort order
func (qb *QueryBuilder) OrderBy(field, direction string) *QueryBuilder {
	qb.sort = &SortOption{
		Field:     field,
		Direction: direction,
	}
	return qb
}

// OrderByAsc sorts by field in ascending order (shorthand)
func (qb *QueryBuilder) OrderByAsc(field string) *QueryBuilder {
	return qb.OrderBy(field, SortAsc)
}

// OrderByDesc sorts by field in descending order (shorthand)
func (qb *QueryBuilder) OrderByDesc(field string) *QueryBuilder {
	return qb.OrderBy(field, SortDesc)
}

// Skip sets the number of documents to skip
func (qb *QueryBuilder) Skip(n int) *QueryBuilder {
	qb.skip = n
	return qb
}

// Limit sets the maximum number of documents to return
func (qb *QueryBuilder) Limit(n int) *QueryBuilder {
	qb.limit = n
	return qb
}

// Take is an alias for Limit (Prisma-style)
func (qb *QueryBuilder) Take(n int) *QueryBuilder {
	return qb.Limit(n)
}

// Build returns the FindOptions for this query
func (qb *QueryBuilder) Build() FindOptions {
	return FindOptions{
		Filters: qb.filters,
		Sort:    qb.sort,
		Skip:    qb.skip,
		Limit:   qb.limit,
	}
}

// String returns a string representation of the query
func (qb *QueryBuilder) String() string {
	return fmt.Sprintf("Query{collection=%s, filters=%d, skip=%d, limit=%d}",
		qb.collection, len(qb.filters), qb.skip, qb.limit)
}

// UpdateBuilder provides a fluent interface for building updates
type UpdateBuilder struct {
	collection  string
	filters     []Query
	setFields   Document
	unsetFields []string
	merge       bool
}

// NewUpdateBuilder creates a new update builder
func NewUpdateBuilder(collection string) *UpdateBuilder {
	return &UpdateBuilder{
		collection:  collection,
		filters:     []Query{},
		setFields:   make(Document),
		unsetFields: []string{},
		merge:       true, // Default to merge mode
	}
}

// Where adds a filter condition
func (ub *UpdateBuilder) Where(field, operator string, value interface{}) *UpdateBuilder {
	ub.filters = append(ub.filters, Query{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return ub
}

// Set sets a field value
func (ub *UpdateBuilder) Set(field string, value interface{}) *UpdateBuilder {
	ub.setFields[field] = value
	return ub
}

// SetMany sets multiple field values
func (ub *UpdateBuilder) SetMany(fields Document) *UpdateBuilder {
	for k, v := range fields {
		ub.setFields[k] = v
	}
	return ub
}

// Unset removes a field
func (ub *UpdateBuilder) Unset(field string) *UpdateBuilder {
	ub.unsetFields = append(ub.unsetFields, field)
	return ub
}

// UnsetMany removes multiple fields
func (ub *UpdateBuilder) UnsetMany(fields []string) *UpdateBuilder {
	ub.unsetFields = append(ub.unsetFields, fields...)
	return ub
}

// Replace sets replace mode (instead of merge)
func (ub *UpdateBuilder) Replace() *UpdateBuilder {
	ub.merge = false
	return ub
}

// Merge sets merge mode (default)
func (ub *UpdateBuilder) Merge() *UpdateBuilder {
	ub.merge = true
	return ub
}

// Build returns the update options
func (ub *UpdateBuilder) Build() (FindOptions, UpdateOptions) {
	findOpts := FindOptions{
		Filters: ub.filters,
	}
	updateOpts := UpdateOptions{
		Set:   ub.setFields,
		Unset: ub.unsetFields,
		Merge: ub.merge,
	}
	return findOpts, updateOpts
}

// DeleteBuilder provides a fluent interface for building deletes
type DeleteBuilder struct {
	collection string
	filters    []Query
}

// NewDeleteBuilder creates a new delete builder
func NewDeleteBuilder(collection string) *DeleteBuilder {
	return &DeleteBuilder{
		collection: collection,
		filters:    []Query{},
	}
}

// Where adds a filter condition
func (db *DeleteBuilder) Where(field, operator string, value interface{}) *DeleteBuilder {
	db.filters = append(db.filters, Query{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return db
}

// Build returns the find options for deletion
func (db *DeleteBuilder) Build() FindOptions {
	return FindOptions{
		Filters: db.filters,
	}
}
