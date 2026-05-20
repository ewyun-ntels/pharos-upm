package query

import (
	"fmt"
)

const (
	OperationWithinRange  = "within_range"
	OperationOutsideRange = "outside_range"
	OperationIsAbove      = "is_above"
	OperationIsBelow      = "is_below"
)

// Threshold defines a comparison condition used for evaluating a numeric value.
// Notes on naming:
// - StartValue is shared between single operations (is_above/is_below) and range operations.
//   - For is_above/is_below, StartValue represents the single Threshold value.
//   - For within_range/outside_range, StartValue is the lower bound of the range.
//
// - EndValue is only used for range operations and represents the upper bound of the range.
// This shared naming avoids duplicating fields (e.g., Value vs StartValue) and keeps the
// decoding keys consistent across operations.
type Threshold struct {
	Id         string            `mapstructure:"id" json:"id"`
	Operation  string            `mapstructure:"condition" json:"condition"`
	StartValue float64           `mapstructure:"start_value" json:"start_value"`
	EndValue   float64           `mapstructure:"end_value" json:"end_value"`
	Severity   string            `mapstructure:"severity" json:"severity"`
	Labels     map[string]string `mapstructure:"labels" json:"labels,omitempty"`
}

func (t *Threshold) Validate() error {
	// Validate supported operations and parameter consistency
	switch t.Operation {
	case OperationIsAbove, OperationIsBelow:
		// Single value comparison; no further validation needed
		return nil
	case OperationWithinRange, OperationOutsideRange:
		// For range operations, ensure start <= end
		if t.StartValue > t.EndValue {
			return fmt.Errorf("invalid range: start_value (%v) must be <= end_value (%v)", t.StartValue, t.EndValue)
		}
		return nil
	default:
		return fmt.Errorf("unknown operation: %s", t.Operation)
	}
}

func (t *Threshold) Check(value float64) bool {
	switch t.Operation {
	case OperationIsAbove:
		return value > t.StartValue
	case OperationIsBelow:
		return value < t.StartValue
	case OperationWithinRange:
		return value >= t.StartValue && value <= t.EndValue
	case OperationOutsideRange:
		return value < t.StartValue || value > t.EndValue
	default:
		return false
	}
}
