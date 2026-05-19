package event_history

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
	alert_common "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/alert/model"
	"ntels.com/pharos/core/pkg/alert/resources"
	rule_common "ntels.com/pharos/core/pkg/alert/rule/common"
	"ntels.com/pharos/core/pkg/common"
)

type Rule struct {
	config          common.Config
	Id              string   `mapstructure:"id,omitempty" json:"id,omitempty"`
	Name            string   `mapstructure:"name" json:"name"`
	Description     string   `mapstructure:"description,omitempty" json:"description,omitempty"`           // alert 발생시 출력할 내용
	RetentionPeriod int      `mapstructure:"retention_period,omitempty" json:"retention_period,omitempty"` // 보관 주기
	Notifications   []string `mapstructure:"notifications,omitempty" json:"notifications,omitempty"`
}

func (r *Rule) Normalize() error {
	return nil
}

func (r *Rule) Load(config common.Config, data map[string]any) error {
	r.config = config

	err := mapstructure.Decode(data, r)
	if err != nil {
		slog.Error("rule load error", "error", err)
		return err
	}
	return nil
}

func (r *Rule) GetId() string {
	return r.Id
}

func (r *Rule) SetId(id string) {
	r.Id = id
}

func (r *Rule) GetName() string {
	return r.Name
}

func (r *Rule) Destroy() error {
	err := rule_common.CronInstance.RemoveJob(r.Id)
	if err != nil {
		slog.Warn("cron remove job failed", "id", r.Id, "error", err)
	}
	return nil
}

func (r *Rule) Execute(_ context.Context) error {
	currentTime := time.Now()
	m := model.Model{
		Config: &r.config,
	}

	alerts, err := m.GetAllAlert(r.Name)
	if err != nil {
		slog.Error("GetAllAlert failed", "error", err)
		return err
	}

	var deleteAlerts []alert_common.Value
	for _, alert := range alerts {
		if alert.UpdatedAt.Before(currentTime.Add(-time.Duration(r.RetentionPeriod) * time.Second)) {
			deleteAlerts = append(deleteAlerts, alert)
		}
	}

	if len(deleteAlerts) == 0 {
		return nil
	}

	err = m.DeleteOldAlerts(&deleteAlerts)
	if err != nil {
		slog.Error("alert delete failed", "error", err)
		return err
	}

	return nil
}

func (r *Rule) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}

	if r.RetentionPeriod <= 0 {
		return errors.New("retention_period must be > 0 seconds")
	}

	return nil
}

func (r *Rule) triggerAlert(alert *alert_common.Value) error {
	if alert.Severity == rule_common.SeverityNormal {
		return nil
	}

	m := model.Model{
		Config: &r.config,
	}

	err := m.UpdateAlert(alert)
	if err != nil {
		slog.Error("db.update failed", "error", err)
		return err
	}

	return nil
}

func (r *Rule) EventHandler(event *resources.Event) error {

	if event.AlertId == "" {
		return errors.New("alert id is required")
	}

	id, uuidErr := uuid.NewRandom()
	if uuidErr != nil {
		slog.Error("uuid generate failed", "error", uuidErr)
		return uuidErr
	}

	currentTime := time.Now().UTC()

	value := alert_common.Value{
		AlertType:   string(alert_common.TypeEventHistory),
		Id:          id.String(),
		Status:      alert_common.StatusEvent,
		Name:        r.Name,
		Labels:      event.Labels,
		Timestamp:   currentTime,
		UpdatedAt:   currentTime,
		CheckTime:   &currentTime,
		Severity:    event.Severity,
		Value:       event.Value,
		Description: event.Description,
		AlertId:     event.AlertId,
	}

	err := value.ConvertLabelsToStringLabels()
	if err != nil {
		slog.Error("history alert convert labels to string labels failed", "error", err)
		return err
	}

	err = r.triggerAlert(&value)
	if err != nil {
		slog.Error("triggerAlert failed", "error", err)
		return err
	}

	n := rule_common.NotificationSender{
		Config:       r.config,
		Destinations: r.Notifications,
	}

	err = n.Send([]alert_common.Value{value})
	if err != nil {
		slog.Error("notificationSender.Send failed", "error", err)
	}

	h := rule_common.History{
		Config: r.config,
	}
	err = h.Send([]alert_common.Value{value}, alert_common.StatusChangeReasonAuto, nil)
	if err != nil {
		slog.Error("alert event history send failed", "error", err)
	}

	return nil
}

func (r *Rule) Run(_ context.Context) error {
	err := rule_common.CronInstance.AddJob("* * * * * *", r)
	if err != nil {
		return err
	}

	return nil
}
