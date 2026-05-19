package plugins

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/core/pkg/plugins/shared"
	sharedRole "ntels.com/pharos/shared/types/role"
)

type Datasource struct {
	model.Datasource

	Provisioning bool `json:"provisioning" db:"-"`

	DataForDB []byte `json:"-" db:"data"`
}

func (datasource *Datasource) GetHandler(c *gin.Context) (int, any) {
	var response any
	var err error

	name := c.Param("name")
	if len(name) != 0 {
		err = datasource.SetFromDB(name)
		response = datasource
	} else {
		response, err = datasource.gets()
	}

	if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, response
}

func (datasource *Datasource) PostHandler(c *gin.Context) (int, any) {
	if err := datasource.setFromReader(c.Request.Body); err != nil {
		return http.StatusBadRequest, external.ErrorResponse{Message: err.Error()}
	}

	if datasource.exist() {
		return http.StatusInternalServerError, external.ErrorResponse{Message: "datasource that already exists"}
	}

	if err := datasource.insert(); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	request := &model.RunStreamRequest{
		Datasource: datasource.Datasource,
		Headers: map[string]string{
			"websocket-endpoints": strings.Join(common.GetWebsocketEndpoints(config), ","),
		},
	}

	response := shared.DatasourceClients.RunStream(request)
	if len(response.Error) != 0 && response.Error != external.ErrorNotSupportedStreamServer.Error() {
		return http.StatusInternalServerError, external.ErrorResponse{Message: response.Error}
	}

	return http.StatusOK, nil
}

func (datasource *Datasource) PutHandler(c *gin.Context) (int, any) {
	name := c.Param("name")

	if provisioningDatasources.Exist(name) {
		return http.StatusBadRequest, external.ErrorResponse{Message: "provisioning datasources cannot be modified"}
	}

	oldDatasource := Datasource{}
	if err := oldDatasource.SetFromDB(name); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if err := oldDatasource.removeDatasource(); err != nil {
		slog.Error("remove datasource error", "error", err.Error())
	}

	if err := datasource.setFromReader(c.Request.Body); err != nil {
		return http.StatusBadRequest, external.ErrorResponse{Message: err.Error()}
	} else if err := datasource.JsonToDB(); err != nil {
		return http.StatusBadRequest, external.ErrorResponse{Message: err.Error()}
	}
	datasource.Name = name

	if err := datasource.update(); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	request := &model.RunStreamRequest{
		Datasource: datasource.Datasource,
		Headers: map[string]string{
			"websocket-endpoints": strings.Join(common.GetWebsocketEndpoints(config), ","),
		},
	}
	response := shared.DatasourceClients.RunStream(request)
	if len(response.Error) != 0 && response.Error != external.ErrorNotSupportedStreamServer.Error() {
		return http.StatusInternalServerError, external.ErrorResponse{Message: response.Error}
	}

	return http.StatusOK, nil
}

func (datasource *Datasource) DeleteHandler(c *gin.Context) (int, any) {
	name := c.Param("name")

	if provisioningDatasources.Exist(name) {
		return http.StatusBadRequest, external.ErrorResponse{Message: "provisioning datasources cannot be deleted"}
	}

	if err := datasource.SetFromDB(name); errors.Is(err, sql.ErrNoRows) {
		return http.StatusOK, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if err := datasource.removeDatasource(); err != nil {
		slog.Error("remove datasource error", "error", err.Error())
	}

	handler := func(db *sqlx.DB) error {
		_, err := db.Exec(`DELETE FROM plugin_datasource WHERE name = $1;`, name)
		return err
	}

	if err := orm.Handler(orm.DriverDefault, &config.Database, handler); errors.Is(err, sql.ErrNoRows) {
		return http.StatusOK, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (datasource *Datasource) exist() bool {
	if provisioningDatasources.Exist(datasource.Name) {
		return true
	}

	handler := func(db *sqlx.DB) error {
		return db.Get(datasource, `SELECT * FROM plugin_datasource WHERE name = $1;`, datasource.Name)
	}

	if err := orm.Handler(orm.DriverDefault, &config.Database, handler); errors.Is(err, nil) {
		return true
	}

	return false
}

func (datasource *Datasource) gets() ([]Datasource, error) {
	datasources, err := datasource.getsDB()
	if err != nil {
		return nil, err
	}

	for _, ds := range provisioningDatasources.GetAll() {
		datasources = append(datasources, *ds)
	}

	return datasources, nil
}

func (datasource *Datasource) getsDB() ([]Datasource, error) {
	var datasources []Datasource

	handler := func(db *sqlx.DB) error {
		return db.Select(&datasources, `SELECT * FROM plugin_datasource;`)
	}

	if err := orm.Handler(orm.DriverDefault, &config.Database, handler); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	for index := range datasources {
		if err := datasources[index].DbToJson(); err != nil {
			return nil, err
		}
	}

	return datasources, nil
}

func (datasource *Datasource) getsDBToMap() (map[string]Datasource, error) {
	datasources, err := datasource.getsDB()
	if err != nil {
		return nil, err
	}

	result := map[string]Datasource{}

	for _, ds := range datasources {
		result[ds.Name] = ds
	}

	return result, nil
}

func (datasource *Datasource) insert() error {
	if err := datasource.validate(); err != nil {
		return err
	}

	if err := datasource.JsonToDB(); err != nil {
		return err
	}

	handler := func(db *sqlx.DB) error {
		_, err := db.NamedExec(`
INSERT INTO plugin_datasource(name, type, data) VALUES
(:name, :type, :data);
`, datasource)
		return err
	}

	return orm.Handler(orm.DriverDefault, &config.Database, handler)
}

func (datasource *Datasource) update() error {
	if err := datasource.validate(); err != nil {
		return err
	}

	handler := func(db *sqlx.DB) error {
		_, err := db.NamedExec(`
UPDATE plugin_datasource SET
    type = :type,
    data = :data
WHERE name = :name;`, datasource)

		return err
	}

	return orm.Handler(orm.DriverDefault, &config.Database, handler)
}

func (datasource *Datasource) SetFromDB(name string) error {
	handler := func(db *sqlx.DB) error {
		return db.Get(datasource, `SELECT * FROM plugin_datasource WHERE name = $1;`, name)
	}

	if err := orm.Handler(orm.DriverDefault, &config.Database, handler); err != nil {
		return err
	}

	if err := datasource.DbToJson(); err != nil {
		return err
	}

	return nil
}

func (datasource *Datasource) setFromReader(reader io.Reader) error {
	if body, err := io.ReadAll(reader); err != nil {
		return err
	} else if err := json.Unmarshal(body, datasource); err != nil {
		return err
	} else if _, exist := shared.PluginsByID[datasource.Type]; !exist {
		return external.ErrorInvalidDatasourceType
	}

	if err := datasource.JsonToDB(); err != nil {
		return err
	}

	return datasource.validate()
}

func (datasource *Datasource) DbToJson() error {
	if err := json.Unmarshal(datasource.DataForDB, &datasource.Data); err != nil {
		return err
	}

	return nil
}

func (datasource *Datasource) JsonToDB() error {
	var err error

	if datasource.DataForDB, err = json.Marshal(datasource.Data); err != nil {
		return err
	}

	return nil
}

func (datasource *Datasource) validate() error {
	if bytes, err := json.Marshal(datasource); err != nil {
		return err
	} else if err := shared.PluginsByID[datasource.Type].ValidateJSONSchemaInstance(string(bytes)); err != nil {
		return err
	}

	return nil
}

func (datasource *Datasource) removeDatasource() error {
	databaseConfig, err := datasource.Datasource.ToDatabaseConfig(config)
	if err != nil {
		return err
	}

	request := &model.RemoveDatasourceRequest{
		Datasource:     datasource.Datasource,
		DatabaseConfig: databaseConfig,
	}

	if err := shared.DatasourceClients.RemoveDatasource(request); err != nil {
		return err
	}

	return nil
}

func pluginsHandler(c *gin.Context) {
	var plugins []*model.Plugin

	for _, plugin := range shared.PluginsByID {
		plugins = append(plugins, plugin)
	}

	c.JSON(http.StatusOK, plugins)
}

func datasourcesHandler(c *gin.Context) {
	datasource := Datasource{}

	switch c.Request.Method {
	case http.MethodGet:
		c.JSON(datasource.GetHandler(c))
	case http.MethodPost:
		c.JSON(datasource.PostHandler(c))
	case http.MethodPut:
		c.JSON(datasource.PutHandler(c))
	case http.MethodDelete:
		c.JSON(datasource.DeleteHandler(c))
	default:
		c.JSON(http.StatusNotImplemented, nil)
	}
}

func dsQueryPostHandler(c *gin.Context) {
	dsQueryRequest := DsQueryRequest{}
	if err := dsQueryRequest.set(c.Request.Body); err != nil {
		c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, DatasourceQuery(c, dsQueryRequest))
}

// RoleGroup represents a group of roles from a single extension
type RoleGroup struct {
	ExtensionID   string                    `json:"extensionId"`
	ExtensionName string                    `json:"extensionName"`
	Roles         []sharedRole.RoleMetadata `json:"roles"`
}

// extensionRolesHandler returns all roles from all loaded extensions
// GET /api/plugins/roles
func extensionRolesHandler(c *gin.Context) {
	var result []RoleGroup

	extensions := extensionRegistry.GetAll()

	for name, ext := range extensions {
		roles := ext.GetRoles()
		if len(roles) > 0 {
			result = append(result, RoleGroup{
				ExtensionID:   name,
				ExtensionName: ext.GetName(),
				Roles:         roles,
			})
		}
	}

	c.JSON(http.StatusOK, result)
}

func RegisterRoutes(_ common.Config, routes gin.IRoutes) {
	routes.GET("", pluginsHandler)

	routes.GET("/datasources", datasourcesHandler)
	routes.GET("/datasources/:name", datasourcesHandler)
	routes.POST("/datasources", datasourcesHandler)
	routes.PUT("/datasources/:name", datasourcesHandler)
	routes.DELETE("/datasources/:name", datasourcesHandler)

	routes.POST("/ds/query", dsQueryPostHandler)

	// Extension roles API
	routes.GET("/roles", extensionRolesHandler)
}
