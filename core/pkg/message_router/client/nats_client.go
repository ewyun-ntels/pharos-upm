package client

import (
	"errors"
	"log/slog"
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/message_router"
)

type NatsClient struct {
	mutex sync.Mutex

	conn *nats.Conn
}

func (c *NatsClient) Connect(config common.Config) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exist := config.Clients[common.ClientTypeNats]; !exist {
		slog.Info("Since the nats client configuration does not exist, no connection is made")
		return nil
	}

	url := config.Clients[common.ClientTypeNats].URL
	if len(url) == 0 {
		slog.Info("Since the nats client URL is empty, no connection is made")
		return nil
	}

	if c.conn != nil && c.conn.IsConnected() {
		return nil
	}

	options := c.getOptions()

	if strings.HasPrefix(url, "tls://") {
		caFile := config.Clients[common.ClientTypeNats].CaFile
		certFile := config.Clients[common.ClientTypeNats].CertFile
		keyFile := config.Clients[common.ClientTypeNats].KeyFile

		options = append(options, nats.RootCAs(caFile))
		options = append(options, nats.ClientCert(certFile, keyFile))
	}

	if conn, err := nats.Connect(url, options...); err != nil {
		return err
	} else {
		c.conn = conn
	}

	if err := c.subscribe(); err != nil {
		return err
	}

	return nil
}

func (c *NatsClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn == nil {
		return nil
	}

	if c.conn.IsClosed() {
		c.conn = nil
		return nil
	}

	c.conn.Close()
	c.conn = nil

	return nil
}

func (c *NatsClient) Request(subjectKey, target string, data []byte) (message_router.Message, error) {
	msg, err := c.getMsg(subjectKey, target, data)
	if err != nil {
		return message_router.Message{}, err
	}

	natsMsg, err := c.conn.RequestMsg(&msg, 5*time.Second)
	if err != nil {
		return message_router.Message{}, err
	}

	// NATS 메시지를 공통 Message 타입으로 변환
	return c.toCommonMessage(natsMsg), nil
}

// toCommonMessage는 nats.Msg를 공통 Message 타입으로 변환합니다
func (c *NatsClient) toCommonMessage(natsMsg *nats.Msg) message_router.Message {
	headers := make(map[string][]string)
	if natsMsg.Header != nil {
		maps.Copy(headers, natsMsg.Header)
	}

	return message_router.Message{
		Data:    natsMsg.Data,
		Subject: natsMsg.Subject,
		Reply:   natsMsg.Reply,
		Headers: headers,
		Raw:     natsMsg, // 원본 메시지 보관
	}
}

func (c *NatsClient) Publish(subjectKey string, data []byte) error {
	if msg, err := c.getMsg(subjectKey, "", data); err != nil {
		return err
	} else if err := c.conn.PublishMsg(&msg); errors.Is(err, nats.ErrInvalidConnection) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

func (c *NatsClient) Subscribe(subject string, handler message_router.MessageHandler) error {
	// 공통 MessageHandler를 NATS MsgHandler로 변환
	natsHandler := func(natsMsg *nats.Msg) {
		// NATS 메시지를 공통 Message로 변환
		msg := c.toCommonMessage(natsMsg)
		// 공통 핸들러 호출
		handler(msg)
	}

	if _, err := c.conn.Subscribe(subject, natsHandler); err != nil {
		return err
	}

	return nil
}

func (c *NatsClient) getOptions() []nats.Option {
	return []nats.Option{
		nats.Name(message_router.GetName()),
		nats.Timeout(10 * time.Second),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(10 * time.Second),

		//nats.NoEcho(),
		//nats.RetryOnFailedConnect(true),
		//nats.CustomReconnectDelay(func(attempts int) time.Duration { return 5 * time.Minute }),

		nats.ClosedHandler(func(conn *nats.Conn) { slog.Info("ClosedHandler call") }),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			if errors.Is(err, nil) {
				slog.Info("DisconnectHandler call")
			} else {
				slog.Info("DisconnectHandler call", "error", err.Error())
			}
		}),
		nats.ConnectHandler(func(conn *nats.Conn) { slog.Info("ConnectHandler call") }),
		nats.ReconnectHandler(func(conn *nats.Conn) { slog.Info("ReconnectHandler call") }),
		nats.ReconnectErrHandler(func(conn *nats.Conn, err error) {
			slog.Info("nats.ReconnectErrHandler call", "ConnectedUrl", conn.ConnectedUrl(), "error", err.Error())
		}),
		nats.DiscoveredServersHandler(func(conn *nats.Conn) {
			slog.Info("DiscoveredServersHandler call", "Servers", conn.Servers(), "DiscoveredServers", conn.DiscoveredServers())
		}),
		nats.ErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
			slog.Info("ErrorHandler call", "error", err.Error(), "Subject", subscription.Subject)
		}),
	}
}

func (c *NatsClient) getMsg(subjectKey, target string, data []byte) (nats.Msg, error) {
	if !message_router.Subjects.Exist(subjectKey) {
		return nats.Msg{}, errors.New("not exist subject")
	}

	return nats.Msg{
		Subject: strings.Replace(message_router.Subjects.Get(subjectKey).RequestSubject, "{{.name}}", target, 1),
		Reply:   strings.Replace(message_router.Subjects.Get(subjectKey).ReplySubject, "{{.name}}", message_router.GetName(), 1),
		Data:    data,
	}, nil
}

func (c *NatsClient) subscribe() error {
	for _, subject := range message_router.Subjects.GetAll() {
		var subscribeSubject string
		var handler message_router.MessageHandler

		subscribeSubject = strings.Replace(subject.ReplySubject, "{{.name}}", message_router.GetName(), 1)
		handler = subject.ReplyHandler

		if len(subscribeSubject) == 0 || handler == nil {
			continue
		}

		if err := c.Subscribe(subscribeSubject, handler); err != nil {
			return err
		}
	}

	if err := c.conn.Flush(); err != nil {
		return err
	}

	return nil
}
