package db

import (
	"fmt"
)

// matchesFilters checks if a document matches all filter conditions
func matchesFilters(doc Document, filters []Query) bool {
	for _, filter := range filters {
		if !matchesFilter(doc, filter) {
			return false
		}
	}
	return true
}

// matchesFilter checks if a document matches a single filter condition
func matchesFilter(doc Document, filter Query) bool {
	val, ok := doc[filter.Field]
	if !ok {
		// Field doesn't exist
		return filter.Operator == OpNe
	}

	switch filter.Operator {
	case OpEq:
		return equal(val, filter.Value)
	case OpNe:
		return !equal(val, filter.Value)
	case OpGt:
		return greaterThan(val, filter.Value)
	case OpGte:
		return greaterThanOrEqual(val, filter.Value)
	case OpLt:
		return lessThan(val, filter.Value)
	case OpLte:
		return lessThanOrEqual(val, filter.Value)
	case OpIn:
		return inArray(val, filter.Value)
	default:
		return false
	}
}

// Comparison functions

func equal(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func greaterThan(a, b interface{}) bool {
	aNum := toFloat64(a)
	bNum := toFloat64(b)
	return aNum > bNum
}

func greaterThanOrEqual(a, b interface{}) bool {
	aNum := toFloat64(a)
	bNum := toFloat64(b)
	return aNum >= bNum
}

func lessThan(a, b interface{}) bool {
	aNum := toFloat64(a)
	bNum := toFloat64(b)
	return aNum < bNum
}

func lessThanOrEqual(a, b interface{}) bool {
	aNum := toFloat64(a)
	bNum := toFloat64(b)
	return aNum <= bNum
}

func inArray(val interface{}, list interface{}) bool {
	switch v := list.(type) {
	case []interface{}:
		for _, item := range v {
			if equal(val, item) {
				return true
			}
		}
	case []string:
		for _, item := range v {
			if equal(val, item) {
				return true
			}
		}
	}
	return false
}

func toFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	default:
		return 0
	}
}

// sortResults sorts documents by the specified field and direction
func sortResults(results []Document, opt *SortOption) {
	// Simple bubble sort (can be optimized with sort.Slice)
	n := len(results)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			a := results[j][opt.Field]
			b := results[j+1][opt.Field]

			shouldSwap := false
			if opt.Direction == SortDesc {
				shouldSwap = lessThan(a, b)
			} else {
				shouldSwap = greaterThan(a, b)
			}

			if shouldSwap {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}
