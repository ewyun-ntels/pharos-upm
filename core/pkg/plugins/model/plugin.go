package model

import (
	"strings"

	"github.com/iancoleman/orderedmap"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"ntels.com/pharos/core/external"
)

type Plugin struct {
	Type     PluginType `json:"type"`
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Settings Settings   `json:"settings"`

	DatasourceClient *DatasourceClient `json:"-"`
	SampleFormData   string            `json:"-"`
}

func (plugin *Plugin) ValidateJSONSchemaInstance(instance string) error {
	if plugin.Settings.JSONSchema == nil {
		return external.ErrorNotExistPluginJSONSchema
	}

	if doc, err := jsonschema.UnmarshalJSON(strings.NewReader(instance)); err != nil {
		return err
	} else if err := plugin.Settings.JSONSchema.Validate(doc); err != nil {
		return err
	}

	return nil
}

type Settings struct {
	JSONSchema        *jsonschema.Schema     `json:"-"`
	JSONSchemaForJson *orderedmap.OrderedMap `json:"jsonSchema"`

	UiSchemaForJson *orderedmap.OrderedMap `json:"uiSchema"`
}

func MakePlugin(pluginType PluginType, id, name, jsonSchema, uiSchema string, datasourceClient *DatasourceClient, sampleFormData string) (*Plugin, error) {
	plugin := &Plugin{
		Type: pluginType,
		ID:   id,
		Name: name,
		Settings: Settings{
			JSONSchemaForJson: orderedmap.New(),
			UiSchemaForJson:   orderedmap.New(),
		},
		DatasourceClient: datasourceClient,
		SampleFormData:   sampleFormData,
	}
	plugin.Settings.JSONSchemaForJson.SetEscapeHTML(false)
	plugin.Settings.UiSchemaForJson.SetEscapeHTML(false)

	if err := plugin.Settings.JSONSchemaForJson.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		return nil, err
	}

	if err := plugin.Settings.UiSchemaForJson.UnmarshalJSON([]byte(uiSchema)); err != nil {
		return nil, err
	}

	compiler := jsonschema.NewCompiler()
	if doc, err := jsonschema.UnmarshalJSON(strings.NewReader(jsonSchema)); err != nil {
		return nil, err
	} else if err := compiler.AddResource(plugin.ID, doc); err != nil {
		return nil, err
	} else if schema, err := compiler.Compile(plugin.ID); err != nil {
		return nil, err
	} else {
		plugin.Settings.JSONSchema = schema
	}

	return plugin, plugin.ValidateJSONSchemaInstance(sampleFormData)
}
