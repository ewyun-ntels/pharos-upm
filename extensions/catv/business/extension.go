package business

import (
	"ntels.com/pharos/core/pkg/common"
	core_plugins "ntels.com/pharos/core/pkg/plugins"
	core_server "ntels.com/pharos/core/pkg/server"
	"ntels.com/pharos/extensions/catv/business/pkg/collect"
	"ntels.com/pharos/extensions/catv/business/pkg/control"
	"ntels.com/pharos/extensions/catv/business/pkg/migrations"
	"ntels.com/pharos/extensions/catv/business/pkg/plugins"
)

// Extension CatvExtension은 Catv 확장 모듈입니다.
type Extension struct{}

func init() {
	// 빌드 시 SITE_MODE=catv가 설정되면 자동으로 등록됩니다.
	core_server.Register(&Extension{})
	core_plugins.RegisterExtension(plugins.NewCatvPluginExtension())
}

func (e *Extension) Name() string {
	return "catv"
}

func (e *Extension) Load(config common.Config) error {
	if err := control.Load(config); err != nil {
		return err
	}
	if err := migrations.Load(config); err != nil {
		return err
	}
	if err := collect.Load(config); err != nil {
		return err
	}

	return nil
}

func (e *Extension) Unload() {
	collect.Unload()
}
