package migrations

import (
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/extensions/catv/business/pkg/migrations/schema"
)

func Load(config common.Config) error {
	schema.Load(config)

	return nil
}
