package etl

import (
	core_command "ntels.com/pharos/core/pkg/command"
	"ntels.com/pharos/extensions/catv/business/pkg/etl/command"
)

func init() {
	core_command.Register("catv-etl", command.Command())
}
