package plugins

import (
	"encoding/gob"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/internal/plugin/datasource/hashicorp"
	"ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/core/pkg/plugins/plugin/datasource/altibase"

	// "ntels.com/pharos/core/pkg/plugins/plugin/datasource/clickhouse" // migrated to extension
	"ntels.com/pharos/core/pkg/plugins/plugin/datasource/elasticsearch"
	http_receiver "ntels.com/pharos/core/pkg/plugins/plugin/datasource/http-receiver"
	"ntels.com/pharos/core/pkg/plugins/plugin/datasource/postgresql"
	"ntels.com/pharos/core/pkg/plugins/plugin/datasource/prometheus"
	"ntels.com/pharos/core/pkg/plugins/plugin/datasource/vertica"
	"ntels.com/pharos/core/pkg/plugins/shared"
)

var once sync.Once
var pluginNameAllowed = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

type Api struct {
	configPath string
	config     common.Config
}

func (a *Api) Init(configPath string, config common.Config) {
	a.configPath = configPath
	a.config = config
}

func (a *Api) Use() bool {
	return true
}

func (a *Api) Load() error {
	config = a.config

	var err error
	once.Do(func() {
		if err = a.pluginLoad(); err != nil {
			return
		}

		if err = a.hashicorpLoad(); err != nil {
			return
		}

		// Load registered extensions BEFORE provisioning
		if err = a.loadExtensions(); err != nil {
			return
		}

		if err = a.provisioningLoad(); err != nil {
			return
		}

		if err = a.runStream(); err != nil {
			return
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Api) Unload() {
	defer shared.HashicorpClients.RemoveAll(func(client *plugin.Client) { client.Kill() })

	datasources, err := (&Datasource{}).gets()
	if err != nil {
		slog.Error("datasource gets error", "error", err.Error())
	}
	for _, datasource := range datasources {
		if err := datasource.removeDatasource(); err != nil {
			slog.Error("datasourceServer DataServer RemoveDatasource error", "error", err)
			continue
		}
	}
}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("", pluginsHandler)

	routes.GET("/datasources", datasourcesHandler)
	routes.GET("/datasources/:name", datasourcesHandler)
	routes.POST("/datasources", datasourcesHandler)
	routes.PUT("/datasources/:name", datasourcesHandler)
	routes.DELETE("/datasources/:name", datasourcesHandler)

	routes.POST("/ds/query", dsQueryPostHandler)

	// Register extension routes
	if routerGroup, ok := routes.(*gin.RouterGroup); ok {
		a.registerExtensionRoutes(routerGroup)
	}
}

func (a *Api) GetRelativePath() string {
	return "/plugins"
}

func (a *Api) pluginLoad() error {
	// Load built-in plugins
	getPlugins := []func(config common.Config) (*model.Plugin, error){
		// clickhouse.GetPlugin, // migrated to extension
		postgresql.GetPlugin,
		altibase.GetPlugin,
		vertica.GetPlugin,
		prometheus.GetPlugin,
		elasticsearch.GetPlugin,

		http_receiver.GetPlugin,
	}

	for _, getPlugin := range getPlugins {
		if p, err := getPlugin(a.config); err != nil {
			return err
		} else {
			shared.PluginsByID[p.ID] = p
			shared.DatasourceClients.Set(p.ID, p.DatasourceClient)
		}
	}

	// Extensions are already loaded in Load() function before provisioning

	return nil
}

func (a *Api) isSafePluginBinaryName(name string) bool {
	if name != filepath.Base(name) {
		return false
	}
	if !pluginNameAllowed.MatchString(name) {
		return false
	}
	return true
}

func (a *Api) hashicorpLoad() error {
	gob.Register(time.Time{})
	gob.Register([]any{})
	gob.Register(map[string]any{})

	var pluginOutput io.Writer = os.Stdout
	if !a.config.Logger.IsConsole {
		pluginOutput = io.Discard
	}
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugins",
		Output: pluginOutput,
		Level:  hclog.Debug,
	})

	pluginSet := plugin.PluginSet{
		hashicorp.PluginTypeDataServer:   &hashicorp.DataServerPlugin{},
		hashicorp.PluginTypeStreamServer: &hashicorp.StreamServerPlugin{},
	}

	dirEntries, err := os.ReadDir(a.config.Plugins.Path)
	if os.IsNotExist(err) {
		slog.Info("plugin path does not exist", "path", a.config.Plugins.Path)
		return nil
	} else if err != nil {
		return err
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}

		binary := dirEntry.Name()

		if !a.isSafePluginBinaryName(binary) {
			slog.Warn("skipping unsafe plugin binary name", "name", binary)
			continue
		}

		// Build absolute, sanitized path and validate it's within the plugins directory
		pluginsDir, err := filepath.EvalSymlinks(filepath.Clean(a.config.Plugins.Path))
		if err != nil {
			return err
		}
		candidate := filepath.Join(pluginsDir, binary)
		absCandidate, err := filepath.EvalSymlinks(filepath.Clean(candidate))
		if err != nil {
			slog.Warn("skipping plugin due to invalid path", "name", binary, "error", err)
			continue
		}
		// Ensure the resolved path is within the plugins directory
		if rel, err := filepath.Rel(pluginsDir, absCandidate); err != nil || strings.HasPrefix(rel, "..") {
			slog.Warn("skipping plugin outside of plugins directory", "name", binary)
			continue
		}
		// Ensure it's a regular file and executable by the current user
		fi, err := os.Stat(absCandidate)
		if err != nil {
			slog.Warn("skipping plugin due to stat error", "name", binary, "error", err)
			continue
		}
		mode := fi.Mode()
		if !mode.IsRegular() || mode&0111 == 0 {
			slog.Warn("skipping non-executable plugin file", "name", binary)
			continue
		}

		cmd := exec.Command(absCandidate)
		// Remove stdout/stderr settings and control only with logger settings.
		// The HashiCorp go-plugin library is responsible for managing the plugin process's input/output,
		// but setting cmd.Stdout and cmd.Stderr in advance causes conflicts.
		/*
			if !config.Logger.IsConsole {
				// Suppress plugin process stdout/stderr when console logging is disabled
				cmd.Stdout = io.Discard
				cmd.Stderr = io.Discard
			}
		*/
		client := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: hashicorp.HandshakeConfig,
			Plugins:         pluginSet,
			Cmd:             cmd,
			Logger:          logger,
		})

		if _, err := client.Client(); err != nil {
			return err
		}

		shared.HashicorpClients.Set(binary, client)
	}

	return nil
}

func (a *Api) provisioningLoad() error {
	datasourcesDB, err := (&Datasource{}).getsDBToMap()
	if err != nil {
		return err
	}

	if a.config.Plugins.Sample {
		for _, p := range shared.PluginsByID {
			datasource := Datasource{}
			if err := json.Unmarshal([]byte(p.SampleFormData), &datasource); err != nil {
				return err
			}

			datasource.Provisioning = true

			provisioningDatasources.Set(datasource.Name, &datasource)
		}
	}

	for _, object := range a.config.Provisioning.Plugins.Datasources {
		datasource := Datasource{}
		if bytes, err := json.Marshal(object); err != nil {
			return err
		} else if err := json.Unmarshal(bytes, &datasource); err != nil {
			return err
		}

		if _, exist := datasourcesDB[datasource.Name]; exist {
			slog.Error("datasource that already exists", "name", datasource.Name)
			continue
		}

		if err := datasource.validate(); err != nil {
			return err
		}

		datasource.Provisioning = true

		provisioningDatasources.Set(datasource.Name, &datasource)
	}

	return nil
}

func (a *Api) runStream() error {
	datasources, err := (&Datasource{}).gets()
	if err != nil {
		return err
	}

	for _, ds := range datasources {
		request := &model.RunStreamRequest{
			Datasource: ds.Datasource,
			Headers: map[string]string{
				"websocket-endpoints": strings.Join(common.GetWebsocketEndpoints(a.config), ","),
			},
		}

		response := shared.DatasourceClients.RunStream(request)
		if len(response.Error) != 0 && response.Error != external.ErrorNotSupportedStreamServer.Error() {
			slog.Error("DatasourceServer RunStream error", "error", response.Error, "datasource", ds.Datasource)
		}
	}

	return nil
}
