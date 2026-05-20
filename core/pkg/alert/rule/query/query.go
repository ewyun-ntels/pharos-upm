package query

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"ntels.com/pharos/core/external/orm"
	alert_common "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/alert/model"
	"ntels.com/pharos/core/pkg/alert/resources"
	rule_common "ntels.com/pharos/core/pkg/alert/rule/common"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins"
)

const CheckTypeLast = "last"

type Rule struct {
	config             common.Config
	Id                 string      `mapstructure:"id,omitempty" json:"id"`
	Name               string      `mapstructure:"name" json:"name"`
	Description        string      `mapstructure:"description,omitempty" json:"description,omitempty"` // alert 발생시 출력할 내용
	Datasource         string      `mapstructure:"datasource" json:"datasource"`
	DatasourceQuery    Datasource  `mapstructure:"datasource_query" json:"datasource_query"`
	Threshold          []Threshold `mapstructure:"threshold" json:"threshold"`
	QueryDelayOffset   int64       `mapstructure:"query_delay_offset,omitempty" json:"query_delay_offset"` // second 단위로 수집
	EvaluationInterval string      `mapstructure:"evaluation_interval,omitempty" json:"evaluation_interval"`
	CheckType          string      `mapstructure:"check_type,omitempty" json:"check_type"`
	Notifications      []string    `mapstructure:"notifications,omitempty" json:"notifications"`
}

func (r *Rule) Normalize() error {
	for i := range r.Threshold {
		if r.Threshold[i].Id == "" {
			r.Threshold[i].Id = uuid.NewString()
		}
	}
	// Default CheckType if not set
	if r.CheckType == "" {
		r.CheckType = CheckTypeLast
	}
	return nil
}

// SetConfig allows external packages (e.g., tests) to set the rule's configuration safely.
func (r *Rule) SetConfig(cfg common.Config) {
	r.config = cfg
}

func evaluationIntervalSeconds(spec string) int64 {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return 0
	}

	durationSpec := spec
	if strings.HasPrefix(spec, "@every ") {
		durationSpec = strings.TrimSpace(strings.TrimPrefix(spec, "@every "))
	}
	if d, err := time.ParseDuration(durationSpec); err == nil && d > 0 {
		return int64(d.Seconds())
	}

	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(spec)
	if err != nil {
		slog.Warn("failed to parse evaluation interval for history dedup", "spec", spec, "error", err)
		return 0
	}

	now := time.Now()
	next := schedule.Next(now)
	afterNext := schedule.Next(next)
	interval := afterNext.Sub(next)
	if interval <= 0 {
		return 0
	}
	return int64(interval.Seconds())
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

func (r *Rule) GetName() string {
	return r.Name
}

func (r *Rule) GetId() string {
	return r.Id
}

func (r *Rule) SetId(id string) {
	r.Id = id
}

func (r *Rule) validateQueryRule() error {
	if r.Datasource == "" {
		return errors.New("datasource is required")
	}

	if err := r.DatasourceQuery.Validate(); err != nil {
		return err
	}

	if r.EvaluationInterval == "" {
		return errors.New("evaluation_interval is required")
	}

	// Multi-threshold mode requires at least one threshold
	if len(r.Threshold) == 0 {
		return errors.New("at least one threshold must be set")
	}
	// Validate each threshold
	for i := range r.Threshold {
		if r.Threshold[i].Id == "" {
			return fmt.Errorf("threshold[%d] id is required", i)
		}
		if err := r.Threshold[i].Validate(); err != nil {
			return fmt.Errorf("threshold[%d] invalid: %w", i, err)
		}
	}

	switch r.CheckType {
	case CheckTypeLast:
		break
	default:
		return fmt.Errorf("unknown check type: %s", r.CheckType)
	}

	return nil
}

func (r *Rule) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	// Normalize defaults before validation
	if err := r.Normalize(); err != nil {
		return err
	}
	if err := r.validateQueryRule(); err != nil {
		return err
	}
	return nil
}

func (r *Rule) decodeRow(row map[string]any) (alertValues []alert_common.Value, err error) {
	// Use common row parser to unify parsing and errors
	ts, decVal, baseLabels, err := ParseRow(row, r.DatasourceQuery.GetTimeLabel(), r.DatasourceQuery.GetVariableLabel())
	if err != nil {
		return nil, err
	}

	// Ensure all thresholds have an id; if any missing, it's a configuration error
	for i := range r.Threshold {
		if r.Threshold[i].Id == "" {
			return nil, fmt.Errorf("threshold[%d] id is required", i)
		}
	}

	// One alert per matching threshold.
	// Thresholds are validated and normalized during rule initialization/configuration,
	// so at this point they are assumed to be valid.
	for _, th := range r.Threshold {
		if th.Check(decVal) {
			// Merge base labels and threshold labels without mutating inputs
			labels := MergeLabels(baseLabels, th.Labels)

			av := alert_common.Value{
				Name:        r.Name,
				Timestamp:   ts,
				Value:       decVal,
				Severity:    th.Severity,
				Labels:      labels,
				Description: r.Description,
			}

			// Build alert identifier using a deterministic JSON string
			if idStr, err := BuildAlertIdString(th, labels); err == nil {
				av.AlertId = idStr
				av.Id = av.AlertId
			} else {
				slog.Error("failed to build alertId", append([]any{slog.Any("error", err)}, AttrsToArgs(LogAttrs(r.Name, &th, labels))...)...)
			}
			alertValues = append(alertValues, av)
		}
	}

	return alertValues, nil
}

func (r *Rule) buildAlert(alert *alert_common.Value, checkTime time.Time) (err error) {
	// IDs (AlertId and Id) are now set during decodeRow or by caller for clears
	// Convert labels map to string form for DB persistence
	err = alert.ConvertLabelsToStringLabels()
	if err != nil {
		slog.Error("alert.ConvertLabelsToStringLabels failed", append([]any{slog.Any("error", err)}, AttrsToArgs(LogAttrs(r.Name, nil, alert.Labels))...)...)
		return err
	}

	alert.UpdatedAt = alert.Timestamp
	alert.CheckTime = &checkTime
	alert.AlertType = string(alert_common.TypeQuery)
	alert.Mask = false

	return
}

func (r *Rule) triggerAlert(alert *alert_common.Value) error {
	// Always persist status changes, regardless of severity
	m := model.Model{
		Config: &r.config,
	}

	err := m.UpdateAlert(alert)
	if err != nil {
		slog.Error("db.update failed", append([]any{slog.Any("error", err)}, AttrsToArgs(LogAttrs(r.Name, nil, alert.Labels))...)...)
		return err
	}

	return nil
}

func (r *Rule) AlertCheck(data *orm.DatabaseResponse) error {
	var lastValue []alert_common.Value

	checkTime := time.Now()

	for _, row := range data.Data {
		switch r.CheckType {
		case CheckTypeLast:
			checkRows, err := r.decodeRow(row)
			if err != nil {
				return err
			}
			if len(checkRows) == 0 {
				// No matches for this row; skip
				continue
			}
			rowTs := checkRows[0].Timestamp
			if len(lastValue) == 0 || SameTimestamp(lastValue[0].Timestamp, rowTs) {
				lastValue = append(lastValue, checkRows...)
			} else if IsAfter(rowTs, lastValue[0].Timestamp) {
				lastValue = append([]alert_common.Value{}, checkRows...)
			}
		default:
			return fmt.Errorf("alert type is not Query(%s)", r.CheckType)
		}
	}

	m := model.Model{
		Config: &r.config,
	}

	alert, err := m.GetAllAlert(r.Name)
	if err != nil {
		slog.Error("GetAllAlert failed", "error", err)
		return err
	}

	// Map existing alerts by AlertId for quick lookup
	alertMap := make(map[string]*alert_common.Value)
	for i := range alert {
		a := &alert[i]
		alertMap[a.AlertId] = a
	}

	var notifyAlerts []alert_common.Value

	// Process current (latest) alerts: set status to alerting and decide notification on change
	for _, value := range lastValue {
		// Build base fields and keys
		err = r.buildAlert(&value, checkTime)
		if err != nil {
			slog.Error("buildAlert failed", "error", err)
			return err
		}

		// Default to alerting for matched thresholds
		value.Status = alert_common.StatusAlerting

		if currentAlert, exists := alertMap[value.AlertId]; exists {
			// Fill previous info
			prevVal := currentAlert.Value
			value.PreviousValue = &prevVal
			value.PreviousSeverity = currentAlert.Severity
			value.PreviousTimestamp = currentAlert.Timestamp

			// Determine if we should notify/history: status change or severity change
			shouldNotify := currentAlert.Status != alert_common.StatusAlerting || currentAlert.Severity != value.Severity
			if shouldNotify {
				notifyAlerts = append(notifyAlerts, value)
			}
			// Remove from map, since it's accounted for
			delete(alertMap, value.AlertId)
		} else {
			// New alert occurrence
			value.PreviousSeverity = rule_common.SeverityNormal
			notifyAlerts = append(notifyAlerts, value)
		}

		// Persist status/alert update
		if err = r.triggerAlert(&value); err != nil {
			slog.Error("triggerAlert failed", "error", err)
			return err
		}
	}

	// Remaining alerts in alertMap were present previously but not in latest results -> clear them and delete from status table
	for _, currentAlert := range alertMap {
		// Reconstruct labels map for building notification/history payload
		_ = currentAlert.ConvertStringLabelsToLabels()
		clearVal := alert_common.Value{
			Name:              r.Name,
			Timestamp:         checkTime,
			Value:             currentAlert.Value,
			Severity:          rule_common.SeverityNormal,
			Status:            alert_common.StatusNormal,
			Labels:            currentAlert.Labels,
			Description:       r.Description,
			PreviousSeverity:  currentAlert.Severity,
			PreviousValue:     &currentAlert.Value,
			PreviousTimestamp: currentAlert.Timestamp,
			AlertId:           currentAlert.AlertId,
			Id:                currentAlert.Id,
		}
		// Build keys for notification/history only (do not persist normal state)
		if err = r.buildAlert(&clearVal, checkTime); err != nil {
			slog.Error("buildAlert(clear) failed", "error", err)
			return err
		}
		// Delete cleared alert from current status table so only active alerts remain
		if derr := m.DeleteAlert(r.Name, currentAlert.AlertId); derr != nil {
			attrs := LogAttrs(r.Name, nil, currentAlert.Labels)
			slog.Error("DeleteAlert failed", append([]any{slog.Any("error", derr), slog.String("alert_id", currentAlert.AlertId)}, AttrsToArgs(attrs)...)...)
			return derr
		}
		notifyAlerts = append(notifyAlerts, clearVal)
	}

	// Send notifications for occurrences and clears
	n := rule_common.NotificationSender{
		Config:       r.config,
		Destinations: r.Notifications,
	}

	if err = n.Send(notifyAlerts); err != nil {
		slog.Error("notificationSender.Send failed", append([]any{slog.Any("error", err)}, AttrsToArgs(LogAttrs(r.Name, nil, nil))...)...)
	}

	// Record history for both occurrences and clears
	h := rule_common.History{
		Config: r.config,
	}
	dedup := &rule_common.HistoryDedup{
		EvaluationIntervalSeconds: evaluationIntervalSeconds(r.EvaluationInterval),
	}
	if err = h.Send(notifyAlerts, alert_common.StatusChangeReasonAuto, nil, dedup); err != nil {
		slog.Error("history.Send failed", append([]any{slog.Any("error", err)}, AttrsToArgs(LogAttrs(r.Name, nil, nil))...)...)
	}

	return nil
}

func (r *Rule) runQuery(ctx context.Context) error {
	query, err := r.DatasourceQuery.RenderQuery()
	if err != nil {
		slog.Error("GetQuery failed", "error", err)
		return err
	}

	if r.QueryDelayOffset > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(r.QueryDelayOffset) * time.Second):
		}
	}

	// Execute via plugins.DatasourceQuery routing by DatasourceName
	// Build a single-query request with synthetic id "q1"
	req := plugins.DsQueryRequest{Queries: []plugins.DsQuery{
		{ID: "q1", DatasourceName: r.Datasource, SQL: string(query), Timeout: 30},
	}}
	resp := plugins.DatasourceQuery(ctx, req)
	res, ok := resp.Results["q1"]
	if !ok {
		// No result returned for this id; treat as transient (e.g., unknown datasource yet)
		slog.Warn("DatasourceQuery returned no result for id (skip)", "id", "q1", "name", r.Name)
		return nil
	}
	if res.Error != "" {
		// If datasource is not provisioned or its type is not supported (no client registered),
		// log and continue so that future runs can succeed once it becomes available.
		if plugins.IsDatasourceUnavailable(res.Error) {
			slog.Warn("DatasourceQuery skipped due to datasource unavailability", "error", res.Error, "name", r.Name, "datasource", r.Datasource)
			return nil
		}
		// Other errors: propagate
		slog.Error("DatasourceQuery error", "error", res.Error)
		return fmt.Errorf("datasource query error: %s", res.Error)
	}
	// res.Frame is a value; take address when calling AlertCheck. No need for nil check.
	if res.Frame.Rows == 0 {
		slog.Debug("since there is no query result, alert check is not performed.", "name", r.Name)
		return nil
	}

	err = r.AlertCheck(&res.Frame)
	if err != nil {
		slog.Error("AlertCheck failed", "error", err)
		return err
	}

	return nil
}

func (r *Rule) EventHandler(_ *resources.Event) error {
	return errors.New("event unsupported")
}

func (r *Rule) Run(_ context.Context) error {
	err := rule_common.CronInstance.AddJob(r.EvaluationInterval, r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Rule) Destroy() error {
	err := rule_common.CronInstance.RemoveJob(r.Id)
	if err != nil {
		slog.Warn("cron remove job failed", "id", r.Id, "error", err)
	}

	return nil
}

func (r *Rule) Execute(ctx context.Context) error {
	return r.runQuery(ctx)
}
