package common

import (
	"context"

	"ntels.com/pharos/core/pkg/alert/resources"
	"ntels.com/pharos/core/pkg/common"
)

const SeverityCritical = "Critical"
const SeverityMajor = "Major"
const SeverityMinor = "Minor"
const SeverityNormal = "Normal"

type Rule interface {
	Load(config common.Config, data map[string]any) error
	Normalize() error
	GetId() string
	SetId(id string)
	GetName() string
	Validate() error
	Run(ctx context.Context) error // cron 실행
	Destroy() error
	Execute(ctx context.Context) error         // 주기적으로 alert check 해야할 경우 사용
	EventHandler(event *resources.Event) error // alert을 외부에서 가져올때 사용
}
