package query

import (
	"encoding/json"
	"maps"
	"testing"
)

// helper to unmarshal alert-id string
type aid struct {
	ThresholdId string            `json:"threshold_id"`
	Condition   string            `json:"condition"`
	Labels      map[string]string `json:"labels"`
}

func TestBuildAlertIdString_DeterministicAndNonMutating(t *testing.T) {
	th := Threshold{Id: "th-1", Operation: OperationIsAbove}

	baseA := map[string]string{"b": "2", "a": "1"}
	baseB := map[string]string{"a": "1"}
	// insert in different order for map nondeterminism simulation
	baseB["b"] = "2"

	origA := make(map[string]string)
	maps.Copy(origA, baseA)

	s1, err := BuildAlertIdString(th, baseA)
	if err != nil {
		t.Fatalf("BuildAlertIdString failed: %v", err)
	}
	s2, err := BuildAlertIdString(th, baseB)
	if err != nil {
		t.Fatalf("BuildAlertIdString failed: %v", err)
	}
	if s1 != s2 {
		t.Fatalf("expected deterministic AlertId strings, got\n%s\n!=\n%s", s1, s2)
	}

	// Ensure input map not mutated by builder
	for k, v := range origA {
		if baseA[k] != v {
			t.Fatalf("base labels mutated for key %q: want %q got %q", k, v, baseA[k])
		}
	}
	if _, ok := baseA["condition"]; ok {
		t.Fatalf("builder must not inject 'condition' into input labels")
	}
	if _, ok := baseA["threshold_id"]; ok {
		t.Fatalf("builder must not inject 'threshold_id' into input labels")
	}

	// Validate contents
	var idObj aid
	if err := json.Unmarshal([]byte(s1), &idObj); err != nil {
		t.Fatalf("failed to unmarshal alert id json: %v", err)
	}
	if idObj.ThresholdId != th.Id {
		t.Fatalf("threshold_id mismatch: want %q got %q", th.Id, idObj.ThresholdId)
	}
	if idObj.Condition != th.Operation {
		t.Fatalf("condition mismatch: want %q got %q", th.Operation, idObj.Condition)
	}
	if idObj.Labels["condition"] != th.Operation {
		t.Fatalf("labels missing condition: %v", idObj.Labels)
	}
	if idObj.Labels["threshold_id"] != th.Id {
		t.Fatalf("labels missing threshold_id: %v", idObj.Labels)
	}
	if idObj.Labels["a"] != "1" || idObj.Labels["b"] != "2" {
		t.Fatalf("labels lost base entries: %v", idObj.Labels)
	}
}
