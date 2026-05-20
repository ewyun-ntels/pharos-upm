package query

import (
	"errors"

	alert_common "ntels.com/pharos/core/internal/query_builder"
)

type Datasource struct {
	Variables     map[string]any `json:"variables,omitempty" mapstructure:"variables"`
	Query         string         `json:"query" mapstructure:"query"`
	TimeLabel     string         `json:"time_label" mapstructure:"time_label"`
	VariableLabel string         `json:"variable_label" mapstructure:"variable_label"`
}

// Validate ensures required fields are present for query execution.
func (d Datasource) Validate() error {
	if d.Query == "" {
		return errors.New("query is required")
	}
	if d.TimeLabel == "" {
		return errors.New("time label is required")
	}
	if d.VariableLabel == "" {
		return errors.New("variable label is required")
	}
	return nil
}

// RenderQuery renders the query with variables using the shared templating helper.
func (d Datasource) RenderQuery() ([]byte, error) {
	return alert_common.GetQuery(d.Query, d.Variables)
}

// GetTimeLabel returns the time label column name.
func (d Datasource) GetTimeLabel() string { return d.TimeLabel }

// GetVariableLabel returns the value column label to evaluate.
func (d Datasource) GetVariableLabel() string { return d.VariableLabel }
