package subjects

import (
	"log/slog"

	"github.com/nats-io/nats.go"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/message_router"
)

func GetAgentPingSubject(config common.Config) message_router.Subject {
	const subject = "agent.ping.{{.name}}"

	return message_router.Subject{
		Key: message_router.AgentPing,

		RequestSubject: subject,
		RequestHandler: func(message message_router.Message) {
			slog.Info("receive message", "Subject", message.Subject, "Reply", message.Reply, "Data", string(message.Data))

			msg, ok := message.Raw.(*nats.Msg)
			if !ok {
				slog.Error("invalid raw message type")
				return
			}

			if err := msg.Respond(msg.Data); err != nil {
				slog.Error("failed to respond to message", "error", err)
			}
		},

		ReplySubject: "",
		ReplyHandler: nil,
	}
}
