package http

import (
	core_internal "ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/server/internal"
)

var Servers = core_internal.NewMap[internal.Server]()

func AddServer(name string, s internal.Server) {
	Servers.Set(name, s)
}
