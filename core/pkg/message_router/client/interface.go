package client

import (
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/message_router"
)

type Client interface {
	Connect(config common.Config) error
	Close() error

	Request(subjectKey, target string, data []byte) (message_router.Message, error)
	Publish(subjectKey string, data []byte) error
	Subscribe(subjectKey string, handler message_router.MessageHandler) error
}
