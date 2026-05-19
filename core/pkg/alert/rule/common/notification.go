package common

import (
	"encoding/json"
	"log/slog"

	alert_common "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/notification"
	notification_common "ntels.com/pharos/core/pkg/notification/common"
	"ntels.com/pharos/core/pkg/websocket"
	"ntels.com/pharos/core/pkg/websocket/centrifuge/node/handler"
)

type NotificationSender struct {
	Config       common.Config
	Destinations []string
}

func (h *NotificationSender) send(alert *alert_common.Value) error {
	for _, name := range h.Destinations {
		value := notification_common.AlertValue{
			Timestamp:   alert.Timestamp,
			UpdatedAt:   alert.UpdatedAt,
			Name:        alert.Name,
			AlertType:   alert.AlertType,
			Description: alert.Description,
			AlertId:     alert.AlertId,
			Status:      alert.Status,
			Value:       alert.Value,
			Severity:    alert.Severity,
			Labels:      alert.Labels,
		}
		err := notification.SendNotification(name, &value)
		if err != nil {
			slog.Error("send notification failed", "name", name, "error", err)
			continue
		}
	}

	return nil
}

func (h *NotificationSender) Send(alerts []alert_common.Value) error {
	if len(alerts) == 0 {
		return nil
	}

	for _, alert := range alerts {
		err := h.send(&alert)
		if err != nil {
			slog.Error("send notification failed", "error", err)
			continue
		}

		bytes, err := json.Marshal(alert)
		if err != nil {
			slog.Error("json Marshal error", "error", err, "alert", alert)
			continue
		}

		for _, n := range websocket.Nodes {
			if err := n.Publish(handler.ChannelAlert, bytes); err != nil {
				slog.Error("publish error", "error", err, "bytes", string(bytes))
				continue
			}
		}
	}

	return nil
}
