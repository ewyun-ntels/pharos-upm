package query

import (
	"testing"
)

func TestDatasource_Validate_MissingFields(t *testing.T) {
	// missing query
	ds := Datasource{TimeLabel: "ts", VariableLabel: "val"}
	if err := ds.Validate(); err == nil || err.Error() != "query is required" {
		t.Fatalf("expected query is required, got %v", err)
	}
	// missing time label
	ds = Datasource{Query: "SELECT 1", VariableLabel: "val"}
	if err := ds.Validate(); err == nil || err.Error() != "time label is required" {
		t.Fatalf("expected time label is required, got %v", err)
	}
	// missing variable label
	ds = Datasource{Query: "SELECT 1", TimeLabel: "ts"}
	if err := ds.Validate(); err == nil || err.Error() != "variable label is required" {
		t.Fatalf("expected variable label is required, got %v", err)
	}
}

func TestDatasource_RenderQuery_NoVars(t *testing.T) {
	ds := Datasource{Query: "SELECT 1", TimeLabel: "ts", VariableLabel: "val"}
	if err := ds.Validate(); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
	b, err := ds.RenderQuery()
	if err != nil {
		t.Fatalf("RenderQuery failed: %v", err)
	}
	if string(b) != "SELECT 1" {
		t.Fatalf("unexpected rendered query: %q", string(b))
	}
}

func TestDatasource_Getters(t *testing.T) {
	ds := Datasource{TimeLabel: "ts", VariableLabel: "val"}
	if ds.GetTimeLabel() != "ts" {
		t.Fatalf("GetTimeLabel mismatch")
	}
	if ds.GetVariableLabel() != "val" {
		t.Fatalf("GetVariableLabel mismatch")
	}
}
