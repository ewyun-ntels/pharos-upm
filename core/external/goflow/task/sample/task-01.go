package sample

import (
	"context"

	"ntels.com/pharos/core/pkg/common"
)

type Task01Task struct {
	configPath string
	config     common.Config
	data       any
}

func (task01Task *Task01Task) GetName() string {
	return "sample-task-01"
}

func (task01Task *Task01Task) SetConfig(configPath string, config common.Config) {
	task01Task.configPath = configPath
	task01Task.config = config
}

func (task01Task *Task01Task) SetData(data any) {
	task01Task.data = data
}

func (task01Task *Task01Task) GetDataFormat() any {
	return map[string]any{"field-01": "value-01"}
}

func (task01Task *Task01Task) Run(_ context.Context) (any, error) {
	return nil, nil
}
