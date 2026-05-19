package server

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"ntels.com/pharos/core/external/goflow"
	"ntels.com/pharos/core/pkg/message_router/client"
)

type LeaderChecker struct {
	wg   sync.WaitGroup
	stop atomic.Bool
}

func (leaderChecker *LeaderChecker) Start(server *NatsServer) {
	leaderChecker.Stop()

	leaderChecker.stop.Store(false)

	leaderChecker.wg.Go(func() {

		for {
			if leaderChecker.stop.Load() {
				break
			}

			if time.Now().Unix()%10 != 0 {
				time.Sleep(1 * time.Second)
				continue
			}
			time.Sleep(1 * time.Second)

			if leaderChecker.isLeader(server) {
				if err := client.GlobalClient.Connect(server.config); err != nil {
					slog.Error("nats client start error", "error", err.Error())
				}

				if err := goflow.GlobalGoflow.Start(server.configPath, server.config); err != nil {
					slog.Error("goflow start error", "error", err.Error())
				}
			} else {
				client.GlobalClient.Close()
				goflow.GlobalGoflow.Stop()
			}
		}
	})
}

func (leaderChecker *LeaderChecker) Stop() {
	defer leaderChecker.wg.Wait()

	leaderChecker.stop.Store(true)
}

func (leaderChecker *LeaderChecker) isLeader(server *NatsServer) bool {
	if server.natsServer == nil {
		return false
	}

	slog.Debug("leader check information",
		"JetStreamEnabled", server.natsServer.JetStreamEnabled(),
		"JetStreamIsClustered", server.natsServer.JetStreamIsClustered(),
		"JetStreamIsLeader", server.natsServer.JetStreamIsLeader())

	if !server.natsServer.JetStreamEnabled() || !server.natsServer.JetStreamIsClustered() {
		return true
	}

	return server.natsServer.JetStreamIsLeader()
}
