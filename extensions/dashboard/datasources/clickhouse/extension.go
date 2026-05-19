package clickhousedatasource

import (
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins"
)

// init automatically registers the ClickHouse extension when the package is imported
func init() {
	// Register the extension with an empty config - will be set properly later
	extension := NewClickHouseExtension(common.Config{})
	plugins.RegisterExtension(extension)
}

type Extension struct{}

func (e *Extension) Name() string {
	return "clickhouse-datasource"
}

func (e *Extension) Load(_ common.Config) error {
	return nil
}

func (e *Extension) Unload() {
	// Cleanup if needed
}
