package collect

import (
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/extensions/catv/business/pkg/collect/transmission"
	"ntels.com/pharos/extensions/catv/business/pkg/collect/weather"
)

func Load(config common.Config) error {
	if err := weather.Load(config); err != nil {
		return err
	}

	if err := transmission.Load(config); err != nil {
		return err
	}

	return nil
}

func Unload() {
	weather.Unload()
	transmission.Unload()
}
