package log

import (
	"ntels.com/pharos/core/pkg/common"
	core_http_api "ntels.com/pharos/core/pkg/server/http/api"
	core_server "ntels.com/pharos/core/pkg/server"
)

// Extension registers the log extension with the PHAROS server.
type Extension struct{}

func init() {
	core_server.Register(&Extension{})
}

func (e *Extension) Name() string {
	return "log"
}

func (e *Extension) Load(_ common.Config) error {
	core_http_api.AddApi(&Api{})
	return nil
}

func (e *Extension) Unload() {}
