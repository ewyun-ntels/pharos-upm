package shared

import (
	"github.com/hashicorp/go-plugin"
	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/plugins/model"
)

var PluginsByID = map[string]*model.Plugin{}
var HashicorpClients = internal.NewMap[*plugin.Client]()
var DatasourceClients = model.DatasourceClients{Map: internal.NewMap[*model.DatasourceClient]()}
