package route

import (
	"strings"

	"github.com/danqzq/rift/internal/stream"
)

// Selector determines if a DataPoint matches routing criteria.
type Selector interface {
	Matches(p stream.DataPoint) bool
	String() string
}

// FieldSelector matches based on a field value.
type FieldSelector struct {
	Field string // field name to match (e.g., "metric", "label")
	Value string // expected value
}

// NewFieldSelector creates a selector that matches field=value.
func NewFieldSelector(field, value string) *FieldSelector {
	return &FieldSelector{
		Field: field,
		Value: value,
	}
}

// Matches checks if the data point's label matches the expected value.
func (f *FieldSelector) Matches(p stream.DataPoint) bool {
	switch f.Field {
	case "label", "metric", "name", "key":
		return p.Label == f.Value
	default:
		return false
	}
}

// String returns the string representation.
func (f *FieldSelector) String() string {
	return f.Field + "=" + f.Value
}

// AlwaysSelector matches all data points.
type AlwaysSelector struct{}

// Matches always returns true.
func (a *AlwaysSelector) Matches(p stream.DataPoint) bool {
	return true
}

// String returns "*".
func (a *AlwaysSelector) String() string {
	return "*"
}

// ParseSelector parses a selector expression.
// Supported formats:
// - "field=value" -> FieldSelector
// - "*" -> AlwaysSelector
func ParseSelector(expr string) Selector {
	expr = strings.TrimSpace(expr)

	if expr == "" || expr == "*" {
		return &AlwaysSelector{}
	}

	// Try parsing field=value
	if strings.Contains(expr, "=") {
		parts := strings.SplitN(expr, "=", 2)
		if len(parts) == 2 {
			field := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			return NewFieldSelector(field, value)
		}
	}

	// Default: treat as label match
	return NewFieldSelector("label", expr)
}
