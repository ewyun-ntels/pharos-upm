package event_status

import (
	"context"
	"errors"
	"log/slog"
	"strings"
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
	config        common.Config
	Id            string   `mapstructure:"id,omitempty" json:"id,omitempty"`
	Name          string   `mapstructure:"name" json:"name"`
	Description   string   `mapstructure:"description,omitempty" json:"description,omitempty"` // alert 발생시 출력할 내용
	Notifications []string `mapstructure:"notifications,omitempty" json:"notifications"`
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
	return nil
}

func (r *Rule) Execute(_ context.Context) error {
	return nil
}

func (r *Rule) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
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

	currentTime := time.Now().UTC()
	value := alert_common.Value{
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
		slog.Error("event alert convert labels to string labels failed", "error", err)
		return err
	}

	m := model.Model{
		Config: &r.config,
	}

	alert, err := m.GetAlert(r.Name, event.AlertId)
	if err != nil {
		slog.Error("GetAlert failed", "error", err)
		return err
	}
	//var isUpdate bool
	var notificationValue []alert_common.Value
	if alert != nil && alert.AlertId != "" {
		if !strings.EqualFold(alert.Severity, value.Severity) {
			// severity가 바뀔경우 이전 status(prevAlert)은 clear
			prevAlert := *alert
			prevAlert.Timestamp = currentTime
			prevAlert.Status = alert_common.StatusNormal
			prevAlert.CheckTime = &currentTime
			prevVal := alert.Value
			prevAlert.PreviousValue = &prevVal

			value.PreviousTimestamp = alert.Timestamp
			value.PreviousSeverity = alert.Severity

			_ = prevAlert.ConvertStringLabelsToLabels()

			notificationValue = append(notificationValue, prevAlert)

			if !strings.EqualFold(value.Severity, rule_common.SeverityNormal) {
				value.Status = alert_common.StatusAlerting
				id, uuidErr := uuid.NewRandom()
				if uuidErr != nil {
					slog.Error("uuid generate failed", "error", uuidErr)
					return uuidErr
				}

				value.Id = id.String()
				notificationValue = append(notificationValue, value)
			}
		} else {
			// alert 유지
			value.Id = alert.Id
			value.PreviousValue = &alert.Value
			value.PreviousSeverity = alert.Severity
			value.Timestamp = alert.Timestamp
		}
	} else if alert == nil && !strings.EqualFold(value.Severity, rule_common.SeverityNormal) {
		// alert 신규 발생
		value.Status = alert_common.StatusAlerting
		id, uuidErr := uuid.NewRandom()
		if uuidErr != nil {
			slog.Error("uuid generate failed", "error", uuidErr)
			return uuidErr
		}

		value.Id = id.String()
		notificationValue = append(notificationValue, value)
	}

	if strings.EqualFold(value.Severity, rule_common.SeverityNormal) {
		err = m.DeleteAlert(r.Name, value.AlertId)
		if err != nil {
			slog.Error("DeleteAlert failed", "error", err)
			return err
		}
	} else {
		err = r.triggerAlert(&value)
		if err != nil {
			slog.Error("triggerAlert failed", "error", err)
			return err
		}
	}

	//if isUpdate {
	if len(notificationValue) > 0 {
		n := rule_common.NotificationSender{
			Config:       r.config,
			Destinations: r.Notifications,
		}

		err = n.Send(notificationValue)
		if err != nil {
			slog.Error("notificationSender.Send failed", "error", err)
		}

		h := rule_common.History{
			Config: r.config,
		}
		err = h.Send(notificationValue, alert_common.StatusChangeReasonAuto, nil)
		if err != nil {
			slog.Error("alert event history send failed", "error", err)
		}
	}

	return nil
}

func (r *Rule) Run(_ context.Context) error {
	return nil
}
