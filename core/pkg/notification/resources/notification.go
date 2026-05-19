package resources

import (
	"encoding/json"
	"log/slog"
	"time"

	"ntels.com/pharos/core/pkg/notification/common"
	"ntels.com/pharos/core/pkg/notification/model"
)

type Rule struct {
	NotificationType common.Type    `json:"notification_type"`
	Rule             map[string]any `json:"rule"`
	Timestamp        time.Time      `json:"timestamp"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (r *Rule) BuildFromDb(m *model.Rule) error {
	var rule map[string]any
	if err := json.Unmarshal([]byte(m.Rule), &rule); err != nil {
		slog.Error("json unmarshal failed", "error", err)
		return err
	}

	r.NotificationType = common.Type(m.NotificationType)
	r.Rule = rule
	r.Timestamp = m.Timestamp.Time
	r.UpdatedAt = m.UpdatedAt.Time

	return nil
}
