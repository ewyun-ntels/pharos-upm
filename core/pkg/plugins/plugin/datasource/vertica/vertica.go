package vertica

import (
	_ "embed"

	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/internal/plugin/datasource"
	"ntels.com/pharos/core/pkg/plugins/internal/plugin/datasource/sql"
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
		DataClient: datasource.GetDataClient(datasource.PluginLocationBuiltIn, &sql.DataClient{Config: config}, config),
	}

	return model.MakePlugin(model.PluginTypeDatasource, "vertica", "Vertica", jsonSchema, uiSchema, datasourceClient, sampleFormData)
}
