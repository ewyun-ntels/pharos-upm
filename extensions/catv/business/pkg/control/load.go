package control

import (
	core_command "ntels.com/pharos/core/pkg/command"
	"ntels.com/pharos/core/pkg/common"
	core_http_api "ntels.com/pharos/core/pkg/server/http/api"
	"ntels.com/pharos/extensions/catv/business/pkg/control/command"
	"ntels.com/pharos/extensions/catv/business/pkg/control/schedules"
	"ntels.com/pharos/extensions/catv/business/pkg/control/topology"
)

func init() {
	core_command.Register("catv-control", command.Command())
}

func Load(config common.Config) error {
	if !config.Catv.Control.Use {
		return nil
	}

	core_http_api.AddApi(&topology.Api{})
	core_http_api.AddApi(&schedules.Api{})

	return nil
}
