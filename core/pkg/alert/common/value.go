package common

import (
	"encoding/json"
	"slices"
	"time"
)

const StatusEvent = "event"
const StatusAlerting = "alerting"
const StatusNormal = "normal"

const StatusChangeReasonAuto = "auto"
const StatusChangeReasonManual = "manual"
const StatusChangeReasonMask = "mask"

type Value struct {
	Id                string            `db:"id" json:"id"` // 시스템 내부적으로 알람을 구분하기 위한 key, history 삭제등에 필요(중복제거 위해)
	Timestamp         time.Time         `db:"timestamp" json:"timestamp"`
	UpdatedAt         time.Time         `db:"updated_at" json:"updated_at"`           // alert 상태가 변경된 시간
	CheckTime         *time.Time        `db:"check_time" json:"check_time,omitempty"` // rule을 실행한 시간
	Name              string            `db:"name" json:"name"`
	Mask              bool              `db:"mask" json:"mask"`
	AlertType         string            `db:"alert_type" json:"alert_type"`
	Description       string            `db:"description" json:"description"`
	AlertId           string            `db:"alert_id" json:"alert_id"`
	Value             float64           `db:"value" json:"value"`
	Severity          string            `db:"severity" json:"severity"`
	Status            string            `db:"status" json:"status"`
	PreviousTimestamp time.Time         `db:"previous_timestamp" json:"previous_timestamp"`
	PreviousSeverity  string            `db:"previous_severity" json:"previous_severity,omitempty"`
	PreviousValue     *float64          `db:"previous_value" json:"previous_value,omitempty"`
	StringLabels      string            `db:"labels" json:"-"`
	Labels            map[string]string `json:"labels,omitempty"`
}

type Key struct {
	Name   string
	Labels []string
}

func (v *Value) ConvertStringLabelsToLabels() error {
	if v.StringLabels == "" {
		v.Labels = nil
		return nil
	}

	labels := make(map[string]string)
	err := json.Unmarshal([]byte(v.StringLabels), &labels)
	if err != nil {
		return err
	}
	v.Labels = labels

	return nil
}

func (v *Value) ConvertLabelsToStringLabels() error {
	if v.Labels == nil {
		v.StringLabels = ""
		return nil
	}

	jsonBytes, err := json.Marshal(v.Labels)
	if err != nil {
		return err
	}

	v.StringLabels = string(jsonBytes)

	return nil
}

// MakeKey alert key 생성
func (v *Value) MakeKey() error {
	key := Key{}
	labels := make([]string, 0, len(v.Labels))
	for k, v := range v.Labels {
		labels = append(labels, k+":"+v)
	}

	slices.Sort(labels)

	key.Name = v.Name
	key.Labels = labels

	jsonBytes, err := json.Marshal(key)
	if err != nil {
		return err
	}

	v.AlertId = string(jsonBytes)

	return nil
}
