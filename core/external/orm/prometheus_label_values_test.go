package orm

import (
	"reflect"
	"testing"
)

func TestParseGrafanaLabelValuesQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantHandled bool
		wantLabel   string
		wantMatches []string
		wantErr     bool
	}{
		{
			name:        "label only",
			query:       "label_values(pod)",
			wantHandled: true,
			wantLabel:   "pod",
		},
		{
			name:        "metric and label",
			query:       "label_values(up, pod)",
			wantHandled: true,
			wantLabel:   "pod",
			wantMatches: []string{"up"},
		},
		{
			name:        "selector with comma",
			query:       `label_values(up{job=~"$job", instance=~"$instance"}, pod)`,
			wantHandled: true,
			wantLabel:   "pod",
			wantMatches: []string{`up{job=~"$job", instance=~"$instance"}`},
		},
		{
			name:        "non label values",
			query:       "up",
			wantHandled: false,
		},
		{
			name:        "too many args",
			query:       "label_values(up, pod, job)",
			wantHandled: true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, handled, err := parseGrafanaLabelValuesQuery(tt.query)
			if handled != tt.wantHandled {
				t.Fatalf("handled = %v, want %v", handled, tt.wantHandled)
			}
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !handled {
				return
			}
			if got.label != tt.wantLabel {
				t.Fatalf("label = %q, want %q", got.label, tt.wantLabel)
			}
			if !reflect.DeepEqual(got.matches, tt.wantMatches) {
				t.Fatalf("matches = %#v, want %#v", got.matches, tt.wantMatches)
			}
		})
	}
}
