package query

import (
	"context"
	"testing"
)

// Ensure runQuery handles an invalid/non-existent datasource gracefully by logging and continuing (no error).
func TestRunQuery_InvalidDatasource_Handled(t *testing.T) {
	r := baseValidRule()
	// Use a datasource name that is not provisioned/registered anywhere
	r.Datasource = "__does_not_exist__"

	// runQuery should not fail; it should log and return nil so the scheduler continues.
	if err := r.runQuery(context.Background()); err != nil {
		t.Fatalf("expected nil error for invalid datasource (skip), got %v", err)
	}
}
