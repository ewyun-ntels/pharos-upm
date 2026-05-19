package task

import (
	"context"

	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/common"
)

// var tasks = map[string]Task{}
var tasks = internal.NewMap[Task]()

type Task interface {
	GetName() string
	SetConfig(configPath string, config common.Config)
	SetData(data any)
	GetDataFormat() any
	Run(context.Context) (any, error)
}

func Add(task Task) {
	tasks.Set(task.GetName(), task)
}

func Exist(name string) bool {
	return tasks.Exist(name)
}

func Get(name string) Task {
	return tasks.Get(name)
}

func Gets() map[string]Task {
	return tasks.GetAll()
}
