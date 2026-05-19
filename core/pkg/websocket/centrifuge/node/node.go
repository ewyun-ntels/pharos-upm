package node

import (
	"context"
	"log/slog"

	"github.com/centrifugal/centrifuge"
	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/websocket/centrifuge/node/handler"
)

type Node struct {
	cn *centrifuge.Node
}

func (node *Node) Run(config common.Config) error {
	if n, err := centrifuge.New(centrifuge.Config{}); err != nil {
		return err
	} else {
		node.cn = n
	}

	node.cn.OnConnecting(handler.OnConnecting)
	node.cn.OnConnect(func(client *centrifuge.Client) {
		slog.Info("OnConnect", "TransportName", client.Transport().Name(), "TransportProtocol", client.Transport().Protocol(), "Info", string(client.Info()))

		h := handler.Handler{
			Config: config,
			Client: client,
			Terminal: handler.Terminal{
				Client: client,
				Pipes:  internal.NewMap[*handler.Pipe](),
			},
		}

		client.OnSubscribe(h.OnSubscribe)
		client.OnUnsubscribe(h.OnUnsubscribe)
		client.OnPublish(h.OnPublish)
		client.OnDisconnect(h.OnDisconnect)
	})

	return node.cn.Run()
}

func (node *Node) Shutdown() error {
	return node.cn.Shutdown(context.Background())
}

func (node *Node) Publish(channel string, data []byte, opts ...centrifuge.PublishOption) error {
	_, err := node.cn.Publish(channel, data, opts...)
	return err
}

func (node *Node) GetCentrifugeNode() *centrifuge.Node {
	return node.cn
}
