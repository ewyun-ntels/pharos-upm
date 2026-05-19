package datasource

import (
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/internal/plugin/datasource/hashicorp"
	"ntels.com/pharos/core/pkg/plugins/model"
)

type PluginLocation string

const (
	PluginLocationBuiltIn  PluginLocation = "built-in"
	PluginLocationExternal PluginLocation = "external"
)

func GetDataClient(pluginLocation PluginLocation, builtInClient model.Data, config common.Config) model.Data {
	switch pluginLocation {
	case PluginLocationBuiltIn:
		return builtInClient
	case PluginLocationExternal:
		return &hashicorp.DataClient{Config: config}
	default:
		return builtInClient
	}
}

func GetStreamClient(pluginLocation PluginLocation, builtInClient model.Stream, config common.Config) model.Stream {
	switch pluginLocation {
	case PluginLocationBuiltIn:
		return builtInClient
	case PluginLocationExternal:
		return &hashicorp.StreamClient{Config: config}
	default:
		return builtInClient
	}
}
