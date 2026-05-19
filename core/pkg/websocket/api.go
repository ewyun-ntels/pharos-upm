package websocket

import (
	"log/slog"
	"net/http"

	"github.com/centrifugal/centrifuge"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ntels.com/pharos/core/pkg/common"
	centrifuge_node "ntels.com/pharos/core/pkg/websocket/centrifuge/node"
)

var Nodes = make(map[string]*centrifuge_node.Node)

type Api struct {
	configPath string
	config     common.Config

	node centrifuge_node.Node
}

func (a *Api) Init(configPath string, config common.Config) {
	a.configPath = configPath
	a.config = config
}

func (a *Api) Use() bool {
	return true
}

func (a *Api) Load() error {
	if err := a.node.Run(a.config); err != nil {
		return err
	}

	Nodes[uuid.New().String()] = &a.node

	return nil
}

func (a *Api) Unload() {
	if err := a.node.Shutdown(); err != nil {
		slog.Error("node Shutdown error", "error", err)
	}
}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("/centrifuge", gin.WrapH(centrifuge.NewWebsocketHandler(a.node.GetCentrifugeNode(), centrifuge.WebsocketConfig{
		CheckOrigin:        func(r *http.Request) bool { return true },
		ReadBufferSize:     1024,
		UseWriteBufferPool: true,
	})))
}

func (a *Api) GetRelativePath() string {
	return "/websocket"
}
