package query_builder

// QuerySpec represents a query specification with variables and metadata
type QuerySpec struct {
	Variables     map[string]any `json:"variables" mapstructure:"variables"`
	Query         string         `json:"query" mapstructure:"query"`
	TimeLabel     string         `json:"time_label" mapstructure:"time_label"`
	VariableLabel string         `json:"variable_label" mapstructure:"variable_label"`
}

// Query represents a datasource query with its specification
type Query struct {
	Datasource string    `json:"datasource"`
	Query      QuerySpec `json:"query" mapstructure:"query"`
}
