package task

import (
	"ntels.com/pharos/core/external/goflow/task/sample"
	"ntels.com/pharos/core/external/goflow/task/sqlite"
)

func init() {
	for _, t := range []Task{
		&sample.Task01Task{},
		&sample.Task02Task{},

		&sqlite.BackupTask{},
		&sqlite.VacuumTask{},
	} {
		tasks.Set(t.GetName(), t)
	}
}
