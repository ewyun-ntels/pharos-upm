package sample

import (
	"context"

	"ntels.com/pharos/core/pkg/common"
)

type Task02Task struct {
	configPath string
	config     common.Config
	data       any
}

func (task02Task *Task02Task) GetName() string {
	return "sample-task-02"
}

func (task02Task *Task02Task) SetConfig(configPath string, config common.Config) {
	task02Task.configPath = configPath
	task02Task.config = config
}

func (task02Task *Task02Task) SetData(data any) {
	task02Task.data = data
}

func (task02Task *Task02Task) GetDataFormat() any {
	return "value"
}

func (task02Task *Task02Task) Run(_ context.Context) (any, error) {
	return nil, nil
}
