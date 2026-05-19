package gnet

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/panjf2000/gnet/v2"
	"ntels.com/pharos/core/pkg/common"
)

type Server struct {
	Protocol     string
	Port         int
	EventHandler gnet.EventHandler
}

func (s *Server) Start(configPath string, config common.Config, logger *slog.Logger) error {
	go func() {
		gnetLogger := NewSlogGnetAdapter(logger)
		if err := gnet.Run(s.EventHandler, s.getAddress(), gnet.WithLogger(gnetLogger)); err != nil {
			logger.Error("Failed to start gnet server", "error", err, "address", s.getAddress())
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//lint:ignore SA1019 Deprecated gnet.Stop is used to maintain compatibility with the existing codebase.
	return gnet.Stop(ctx, s.getAddress())
}

func (s *Server) getAddress() string {
	return fmt.Sprintf("%s://:%d", s.Protocol, s.Port)
}
