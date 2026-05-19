package main

import (
	"net/http"

	"github.com/hashicorp/go-plugin"
	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/plugins/internal/plugin/datasource/hashicorp"
	http_receiver "ntels.com/pharos/core/pkg/plugins/plugin/datasource/http-receiver"
)

func main() {
	pluginSet := plugin.PluginSet{
		hashicorp.PluginTypeStreamServer: &hashicorp.StreamServerPlugin{StreamServer: &http_receiver.StreamClient{Servers: internal.NewMap[*http.Server]()}},
	}

	hashicorp.Main(pluginSet)
}
