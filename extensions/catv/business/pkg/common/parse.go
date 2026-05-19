package common

import (
	"fmt"
	"time"

	"ntels.com/pharos/core/external/orm"
)

func ParseKSTDatetime(timeStr string) (*orm.Datetime, error) {
	const location = "Asia/Seoul"
	const layout = "2006/01/02 15:04"

	if timeLocation, err := time.LoadLocation(location); err != nil {
		return nil, fmt.Errorf("failed to load time location: %w, location: %s", err, location)
	} else if t, err := time.ParseInLocation(layout, timeStr, timeLocation); err != nil {
		return nil, fmt.Errorf("failed to parse time: %w, timeStr: %s, layout: %s, location: %s", err, timeStr, layout, location)
	} else {
		return &orm.Datetime{Time: t.UTC()}, nil
	}
}

func ParseRunningTimeSec(runningTimeStr string) (uint64, error) {
	var day, hour, min, sec uint64
	_, err := fmt.Sscanf(runningTimeStr, "%dday %dhour %dmin %dsec", &day, &hour, &min, &sec)
	if err != nil {
		return 0, fmt.Errorf("failed to parse running time: %w, runningTimeStr: %s", err, runningTimeStr)
	}
	totalSec := day*24*3600 + hour*3600 + min*60 + sec
	return totalSec, nil
}
