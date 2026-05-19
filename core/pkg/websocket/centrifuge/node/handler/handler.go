package handler

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/centrifugal/centrifuge"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins"
	"ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/core/pkg/plugins/shared"
)

const (
	ChannelAlert    = "alert"
	ChannelTerminal = "terminal"

	ChannelPluginHttpReceiver = "plugin-http-receiver"
)

type Handler struct {
	Config common.Config
	Client *centrifuge.Client

	Terminal Terminal
}

func (handler *Handler) getDatasource(channel string) (plugins.Datasource, error) {
	datasource := plugins.Datasource{}

	split := strings.Split(channel, ":")
	if len(split) < 2 {
		return datasource, external.ErrorInvalidChannel
	}
	datasourceID := split[1]

	if err := datasource.SetFromDB(datasourceID); err != nil {
		return datasource, err
	}

	return datasource, nil
}

func (handler *Handler) OnSubscribe(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
	slog.Info("OnSubscribe", "Channel", e.Channel, "Data", string(e.Data))

	switch strings.Split(e.Channel, ":")[0] {
	case ChannelPluginHttpReceiver:
		datasource, err := handler.getDatasource(e.Channel)
		if err != nil {
			cb(centrifuge.SubscribeReply{}, external.ErrorInvalidDatasourceID)
			return
		}

		request := &model.SubscribeStreamRequest{
			Datasource: datasource.Datasource,
			Client:     handler.Client,

			Path: e.Channel,
		}

		response := shared.DatasourceClients.SubscribeStream(request)
		if len(response.Error) != 0 {
			cb(centrifuge.SubscribeReply{}, errors.New(response.Error))
			return
		}
	case ChannelTerminal:
		handler.Terminal.OnSubscribe(e, cb)
		return
	}

	cb(centrifuge.SubscribeReply{}, nil)
}

func (handler *Handler) OnUnsubscribe(e centrifuge.UnsubscribeEvent) {
	slog.Info("OnUnsubscribe", "Channel", e.Channel)

	switch strings.Split(e.Channel, ":")[0] {
	case ChannelPluginHttpReceiver:
		datasource, err := handler.getDatasource(e.Channel)
		if err != nil {
			return
		}

		request := &model.UnsubscribeStreamRequest{
			Datasource: datasource.Datasource,

			Path: e.Channel,
		}

		err = shared.DatasourceClients.UnsubscribeStream(request)
		if err != nil {
			slog.Error("Unsubscribe error", "error", err)
			return
		}
	case ChannelTerminal:
		handler.Terminal.OnUnsubscribe(e)
	}
}

func (handler *Handler) OnPublish(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
	slog.Info("OnPublish", "Channel", e.Channel, "Data", string(e.Data))

	switch strings.Split(e.Channel, ":")[0] {
	case ChannelTerminal:
		handler.Terminal.OnPublish(e, cb)
		return
	default:
		cb(centrifuge.PublishReply{}, nil)
	}
}

func (handler *Handler) OnDisconnect(e centrifuge.DisconnectEvent) {
	slog.Info("OnDisconnect", "Code", e.Disconnect.Code, "Reason", e.Disconnect.Reason)
}

func OnConnecting(_ context.Context, e centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
	slog.Info("OnConnecting", "Name", e.Name, "TransportName", e.Transport.Name(), "TransportProtocol", e.Transport.Protocol(), "Data", string(e.Data), "Token", e.Token)

	return centrifuge.ConnectReply{
		Credentials: &centrifuge.Credentials{
			UserID:   "tarzan",
			ExpireAt: 0,
		},
	}, nil
}
