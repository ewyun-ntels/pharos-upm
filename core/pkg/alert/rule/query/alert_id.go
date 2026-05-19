package query

import (
	"bytes"
	"encoding/json"
	"maps"
	"sort"
)

type alertId struct {
	ThresholdId string            `json:"threshold_id"`
	Condition   string            `json:"condition"`
	Labels      map[string]string `json:"labels"`
}

// newAlertId builds an alertId from a threshold and a set of base labels.
// It copies the labels and injects threshold metadata (condition, threshold_id)
// without mutating the original labels map.
func newAlertId(th Threshold, baseLabels map[string]string) alertId {
	labels := make(map[string]string, len(baseLabels)+2)
	maps.Copy(labels, baseLabels)
	if th.Operation != "" {
		labels["condition"] = th.Operation
	}
	if th.Id != "" {
		labels["threshold_id"] = th.Id
	}
	return alertId{
		ThresholdId: th.Id,
		Condition:   th.Operation,
		Labels:      labels,
	}
}

// BuildAlertIdString creates a deterministic JSON string for the alert identifier.
// It normalizes the Labels by writing keys in sorted order to ensure the
// resulting string is stable (avoids map iteration randomness).
func BuildAlertIdString(th Threshold, baseLabels map[string]string) (string, error) {
	id := newAlertId(th, baseLabels)
	b, err := json.Marshal(id)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MarshalJSON implements a deterministic JSON marshaler that sorts label keys.
func (a alertId) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(`{"threshold_id":`)
	thBytes, err := json.Marshal(a.ThresholdId)
	if err != nil {
		return nil, err
	}
	buf.Write(thBytes)

	buf.WriteString(`,"condition":`)
	condBytes, err := json.Marshal(a.Condition)
	if err != nil {
		return nil, err
	}
	buf.Write(condBytes)

	buf.WriteString(`,"labels":{`)
	keys := make([]string, 0, len(a.Labels))
	for k := range a.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		kb, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		vb, err := json.Marshal(a.Labels[k])
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(vb)
		if i < len(keys)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteString("}")
	buf.WriteByte('}')

	return buf.Bytes(), nil
}
