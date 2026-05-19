package service

import (
	"log/slog"

	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/message_router"
	"ntels.com/pharos/core/pkg/message_router/client"
	"ntels.com/pharos/extensions/catv/business/pkg/etl/subjects"
)

func NewService(config common.Config) (*Service, error) {
	return &Service{config: config}, nil
}

type Service struct {
	config     common.Config
	natsClient client.NatsClient
}

func (s *Service) Start() error {
	// 연결전에 Subject 등록을 먼저 해야 함
	subject := subjects.GetElasticsearchIndexSubject(s.config)
	message_router.Subjects.Set(subject.Key, subject)

	if err := s.natsClient.Connect(s.config); err != nil {
		slog.Error("nats client start error", "error", err)
		return err
	}

	return nil
}

func (s *Service) Stop() error {
	if err := s.natsClient.Close(); err != nil {
		slog.Error("nats client stop error", "error", err)
		return err
	}

	return nil
}
