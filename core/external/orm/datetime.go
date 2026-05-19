package orm

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Datetime struct {
	time.Time
}

func (mt *Datetime) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		mt.Time = v
		return nil
	case string:
		t, err := time.Parse(time.DateTime, v)
		if err != nil {
			return err
		}
		mt.Time = t
		return nil
	case []byte:
		t, err := time.Parse(time.DateTime, string(v))
		if err != nil {
			return err
		}
		mt.Time = t
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into Datetime", value)
	}
}

func (mt Datetime) Value() (driver.Value, error) {
	return mt.Time.UTC().Format(time.DateTime), nil
}
