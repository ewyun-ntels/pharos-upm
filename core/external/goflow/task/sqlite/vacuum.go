package sqlite

import (
	"context"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

type VacuumTask struct {
	configPath string
	config     common.Config
	data       any
}

func (vacuumTask *VacuumTask) GetName() string {
	return "sqlite-vacuum"
}

func (vacuumTask *VacuumTask) SetConfig(configPath string, config common.Config) {
	vacuumTask.configPath = configPath
	vacuumTask.config = config
}

func (vacuumTask *VacuumTask) SetData(data any) {
	vacuumTask.data = data
}

func (vacuumTask *VacuumTask) GetDataFormat() any {
	return ``
}

func (vacuumTask *VacuumTask) Run(_ context.Context) (any, error) {
	handler := func(db *sqlx.DB) error {
		_, err := db.Exec(`VACUUM;`)
		return err
	}

	return nil, orm.Handler(orm.DriverDefault, &vacuumTask.config.Database, handler)
}
