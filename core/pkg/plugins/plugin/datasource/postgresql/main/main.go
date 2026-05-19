package main

import (
	"github.com/hashicorp/go-plugin"
	"ntels.com/pharos/core/pkg/plugins/internal/plugin/datasource/hashicorp"
	"ntels.com/pharos/core/pkg/plugins/internal/plugin/datasource/sql"
)

func main() {
	pluginSet := plugin.PluginSet{
		hashicorp.PluginTypeDataServer: &hashicorp.DataServerPlugin{DataServer: &sql.DataClient{}},
	}

	hashicorp.Main(pluginSet)
}
