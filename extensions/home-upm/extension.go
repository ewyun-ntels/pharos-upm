package homeupm

import (
	"ntels.com/pharos/core/pkg/common"
	core_server "ntels.com/pharos/core/pkg/server"
	core_http_api "ntels.com/pharos/core/pkg/server/http/api"
)

// Extension registers the home-upm extension with the PHAROS server.
type Extension struct{}

func init() {
	core_server.Register(&Extension{})
}

func (e *Extension) Name() string {
	return "home-upm"
}

func (e *Extension) Load(_ common.Config) error {
	core_http_api.AddApi(&Api{})
	return nil
}

func (e *Extension) Unload() {}
