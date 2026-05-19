package server

import (
	"log/slog"

	"ntels.com/pharos/core/pkg/common"
)

type Server interface {
	Start(configPath string, config common.Config, logger *slog.Logger) error
	Stop() error
}

func NewServer() Server {
	return &NatsServer{}
}
