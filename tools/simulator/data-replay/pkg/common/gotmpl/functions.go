package gotmpl

import "time"

func now() time.Time {
	return time.Now()
}

func location(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}
