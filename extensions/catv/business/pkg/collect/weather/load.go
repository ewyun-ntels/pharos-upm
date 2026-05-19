package weather

import (
	"log/slog"

	core_command "ntels.com/pharos/core/pkg/command"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/extensions/catv/business/pkg/collect/weather/command"
	"ntels.com/pharos/extensions/catv/business/pkg/collect/weather/service"
)

var GlobalService *service.Service

func init() {
	core_command.Register("catv-collect-weather", command.Command())
}

func Load(config common.Config) error {
	if !config.Catv.Collect.Weather.Use {
		return nil
	}

	if service, err := service.NewService(config); err != nil {
		return err
	} else {
		GlobalService = service
	}

	if err := GlobalService.Start(); err != nil {
		return err
	}

	return nil
}

func Unload() {
	if GlobalService != nil {
		if err := GlobalService.Stop(); err != nil {
			slog.Error("Failed to stop weather service", "error", err)
		}
		GlobalService = nil
	}
}
