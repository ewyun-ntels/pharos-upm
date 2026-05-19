package resources

import (
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"ntels.com/pharos/core/internal/query_builder"
	"ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/alert/model"
)

type Rule struct {
	AlertType common.Type    `json:"alert_type"`
	Rule      map[string]any `json:"rule"`
	Status    []common.Value `json:"status,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	UpdateAt  time.Time      `json:"update_at"`
}

func (r *Rule) BuildFromDb(m *model.Rule, v []common.Value) error {
	var rule map[string]any
	if err := json.Unmarshal([]byte(m.Rule), &rule); err != nil {
		slog.Error("json unmarshal failed", "error", err)
		return err
	}

	r.AlertType = common.Type(m.AlertType)
	r.Rule = rule
	r.Status = v
	r.Timestamp = m.Timestamp.Time
	r.UpdateAt = m.UpdatedAt.Time

	return nil
}

// QuerySpec and Query are now imported from query_builder package
type QuerySpec = query_builder.QuerySpec
type Query = query_builder.Query

type Event struct {
	AlertId     string            `json:"alert_id"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`
	Value       float64           `json:"value"`
	Labels      map[string]string `json:"labels"`
}

func (e *Event) Validate() error {
	if e.AlertId == "" {
		return errors.New("alert_id is required")
	}

	if e.Severity == "" {
		return errors.New("severity is required")
	}

	return nil
}
