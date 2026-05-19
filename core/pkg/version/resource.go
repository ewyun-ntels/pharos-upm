package version

import (
	"time"
)

type Response struct {
	Module    string    `json:"module"`
	Version   string    `json:"version"`
	Commit    string    `json:"commit"`
	BuildTime time.Time `json:"build_time"`
	UpdatedAt time.Time `json:"updated_at"`
}
