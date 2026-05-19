package tcp

import (
	"github.com/panjf2000/gnet/v2"
	core_internal "ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/server/internal"
	gnet_internal "ntels.com/pharos/core/pkg/server/internal/gnet"
)

var Servers = core_internal.NewMap[internal.Server]()

func AddGnetServer(name string, port int, eventHandler gnet.EventHandler) {
	Servers.Set(name, &gnet_internal.Server{
		Protocol:     "tcp",
		Port:         port,
		EventHandler: eventHandler,
	})
}
