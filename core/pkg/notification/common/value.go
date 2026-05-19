package common

import "time"

// AlertValue alert으로 부터 데이터를 수신하기 위해 사용
type AlertValue struct {
	Timestamp   time.Time         `json:"timestamp"`
	UpdatedAt   time.Time         `json:"updated_at"` // alert 상태가 변경된 시간
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	AlertType   string            `json:"alert_type"`
	Description string            `json:"description"`
	AlertId     string            `json:"alert_id"`
	Value       float64           `json:"value"`
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels,omitempty"`
}
