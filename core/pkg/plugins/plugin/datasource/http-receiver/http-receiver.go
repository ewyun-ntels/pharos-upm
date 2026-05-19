package http_receiver

import (
	_ "embed"
	"net/http"

	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/internal/plugin/datasource"
	"ntels.com/pharos/core/pkg/plugins/model"
)

//go:embed schema/json.json
var jsonSchema string

//go:embed schema/ui.json
var uiSchema string

//go:embed schema/sample-form-data.json
var sampleFormData string

func GetPlugin(config common.Config) (*model.Plugin, error) {
	datasourceClient := &model.DatasourceClient{
		StreamClient: datasource.GetStreamClient(datasource.PluginLocationBuiltIn, &StreamClient{Servers: internal.NewMap[*http.Server]()}, config),
	}

	return model.MakePlugin(model.PluginTypeDatasource, "http-receiver", "HTTP Receiver", jsonSchema, uiSchema, datasourceClient, sampleFormData)
}
