package hashicorp

import (
	"encoding/gob"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

func Main(pluginSet plugin.PluginSet) {
	gob.Register(time.Time{})
	gob.Register([]any{})
	gob.Register(map[string]any{})

	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	logger.Info("process start")
	defer logger.Info("process end")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         pluginSet,
	})
}
