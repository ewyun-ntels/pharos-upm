package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"ntels.com/pharos/core/internal/cert"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/message_router"
)

type NatsServer struct {
	mutex sync.Mutex

	configPath string
	config     common.Config

	natsServer    *server.Server
	leaderChecker LeaderChecker
}

func (s *NatsServer) Start(configPath string, config common.Config, logger *slog.Logger) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if logger == nil {
		logger = slog.Default()
	}

	if s.natsServer != nil && s.natsServer.Running() {
		logger.Info("nats server already running")
		return nil
	}

	s.configPath = configPath
	s.config = config

	if _, exist := s.config.Servers[common.ServerTypeNats]; !exist {
		logger.Info("nats server config not exist")
		return nil
	}

	port := s.config.Servers[common.ServerTypeNats].Port
	if port == 0 {
		logger.Info(fmt.Sprintf("Since the nats server port is empty, set it to the default port (%d)", server.DEFAULT_PORT))
		port = server.DEFAULT_PORT
	}

	clusterOpts, err := s.getClusterOpts()
	if err != nil {
		return err
	}

	routesStr := strings.Join(s.config.Nats.Routes, ",")

	options := server.Options{
		ServerName: message_router.GetName(),
		Host:       server.DEFAULT_HOST,
		Port:       s.config.Servers[common.ServerTypeNats].Port,
		NoSigs:     true,

		MaxPayload: s.config.Nats.MaxPayload,

		Cluster: clusterOpts,

		JetStream:          s.config.Nats.JetStream,
		JetStreamMaxMemory: s.config.Nats.JetStreamMaxMemory,
		JetStreamMaxStore:  s.config.Nats.JetStreamMaxStore,
		StoreDir:           s.config.Nats.StoreDir,

		Routes:    server.RoutesFromStr(routesStr),
		RoutesStr: routesStr,
	}

	switch s.config.Serve.NatsSchema {
	case common.SchemaNats:
	case common.SchemaTls:
		if _, err := cert.SetCertification(s.config.Servers[common.ServerTypeNats]); err != nil {
			return err
		}

		caFile, certFile, keyFile := cert.GetCertificationFilePath(s.config.Servers[common.ServerTypeNats])

		tlsConfigOpts := &server.TLSConfigOpts{
			CertFile: certFile,
			KeyFile:  keyFile,
			CaFile:   caFile,
			Verify:   s.config.Servers[common.ServerTypeNats].Cert.InsecureSkipVerify,
			Insecure: true,
			Timeout:  10,
		}

		tlsConfig, err := server.GenTLSConfig(tlsConfigOpts)
		if err != nil {
			return err
		}

		options.TLSConfig = tlsConfig
		options.Cluster.TLSConfig = tlsConfig
	}

	if server, err := server.NewServer(&options); err != nil {
		return err
	} else {
		s.natsServer = server
	}

	s.natsServer.SetLogger(NewSlogNatsAdapter(logger), s.config.Nats.Log.Debug, s.config.Nats.Log.Trace)
	s.natsServer.Start()

	if !s.natsServer.ReadyForConnections(10 * time.Second) {
		return errors.New("nats server ReadyForConnections false")
	}

	s.leaderChecker.Start(s)

	return nil
}

func (s *NatsServer) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.natsServer == nil {
		return nil
	}

	if !s.natsServer.Running() {
		return nil
	}

	s.leaderChecker.Stop()

	s.natsServer.Shutdown()
	s.natsServer.WaitForShutdown()

	s.natsServer = nil

	return nil
}

func (s *NatsServer) getClusterOpts() (server.ClusterOpts, error) {
	listenStr := s.config.Nats.Cluster.Listen
	if len(listenStr) == 0 {
		return server.ClusterOpts{}, nil
	}

	clusterURL, err := url.Parse(listenStr)
	if err != nil {
		return server.ClusterOpts{}, err
	}
	clusterHost, clusterPortStr, err := net.SplitHostPort(clusterURL.Host)
	if err != nil {
		return server.ClusterOpts{}, err
	}
	clusterPort, err := strconv.Atoi(clusterPortStr)
	if err != nil {
		return server.ClusterOpts{}, err
	}

	return server.ClusterOpts{
		Name:      s.config.Nats.Cluster.Name,
		Host:      clusterHost,
		Port:      clusterPort,
		ListenStr: listenStr,

		Compression: server.CompressionOpts{
			Mode: server.CompressionS2Auto,
		},
	}, nil
}
