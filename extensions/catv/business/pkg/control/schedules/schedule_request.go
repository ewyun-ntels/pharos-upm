package schedules

import (
	"encoding/json"
	"io"

	"ntels.com/pharos/core/external/orm"
)

type ScheduleRequest struct {
	Name string `json:"name" db:"name"`

	ScheduleType       string        `json:"schedule_type" db:"schedule_type"`                         // Execution type: immediately, once, repeat (for future use)
	ScheduleSpecOnce   *orm.Datetime `json:"schedule_spec_once,omitempty" db:"schedule_spec_once"`     // Specific time for one-time execution (required if schedule_type is "once")
	ScheduleSpecRepeat *string       `json:"schedule_spec_repeat,omitempty" db:"schedule_spec_repeat"` // Cron expression for repeated execution (required if schedule_type is "repeat")

	Sos         []string `json:"sos" db:"sos"`
	L3s         []string `json:"l3s" db:"l3s"`
	Cells       []string `json:"cells" db:"cells"`
	Settopboxes []string `json:"settopboxes" db:"settopboxes"`

	WorkType  string  `json:"work_type" db:"work_type"`
	WorkValue *string `json:"work_value,omitempty" db:"work_value"`

	AreaType string   `json:"area_type"` // Device selection type
	AreaIDs  []string `json:"area_ids"`  // List of identifiers (empty for type="all")
}

func (r *ScheduleRequest) serFromBody(body io.ReadCloser) error {
	decoder := json.NewDecoder(io.LimitReader(body, MaxRequestBodySize))
	decoder.DisallowUnknownFields()

	return decoder.Decode(r)
}
