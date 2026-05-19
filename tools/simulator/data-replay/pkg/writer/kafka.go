package writer

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type KafkaWriter struct {
	BootstrapServers  []string            `yaml:"bootstrap-servers"`
	Topic             string              `yaml:"topic"`
	RetryMax          int                 `yaml:"retry-max"`
	RequiredAcks      sarama.RequiredAcks `yaml:"required-acks"`
	RequestTimeout    time.Duration       `yaml:"request-timeout"`
	ChannelBufferSize int                 `yaml:"channel-buffer-size"`
	FlushMaxMessages  int                 `yaml:"flush-max-messages"`
	FlushFrequency    time.Duration       `yaml:"flush-frequency"`

	// Optional TLS configuration
	TLSEnabled  bool   `yaml:"tls-enabled"`
	TLSCertFile string `yaml:"tls-cert-file"`

	client   sarama.Client
	producer sarama.SyncProducer

	buffer chan *sarama.ProducerMessage
}

func (w *KafkaWriter) init(config yaml.Node) error {
	if w == nil {
		return errors.New("nil pointer")
	}

	var err error
	err = config.Decode(w)
	if err != nil {
		return errors.Wrap(err, "decode yaml error")
	}

	if w.Topic == "" {
		return errors.New("topic is required")
	}

	if len(w.BootstrapServers) < 1 {
		return errors.New("bootstrap server list is empty")
	}

	if w.RetryMax < 0 {
		w.RetryMax = 0
	}

	if w.RequestTimeout <= 0 {
		w.RequestTimeout = 10 * time.Second
	}

	conf := sarama.NewConfig()
	conf.Net.DialTimeout = 500 * time.Millisecond
	conf.Producer.Compression = sarama.CompressionSnappy
	conf.Producer.MaxMessageBytes = 104857600
	conf.Producer.Return.Successes = true
	//conf.Producer.Return.Errors = false
	conf.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	conf.Producer.Retry.Max = w.RetryMax
	conf.Producer.RequiredAcks = w.RequiredAcks
	conf.Producer.Timeout = w.RequestTimeout
	conf.ChannelBufferSize = w.ChannelBufferSize

	conf.Producer.Flush.MaxMessages = w.FlushMaxMessages
	conf.Producer.Flush.Frequency = w.FlushFrequency

	if w.TLSEnabled {
		conf.Net.TLS.Enable = true

		// Get tle.Certificate from filename
		certData, err := os.ReadFile(w.TLSCertFile)
		if err != nil {
			return errors.Wrap(err, "failed to read certificate file")
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(certData) {
			return errors.New("failed to append certificate to pool")
		}
		conf.Net.TLS.Config = &tls.Config{
			RootCAs: certPool,
		}
	}

	w.client, err = sarama.NewClient(w.BootstrapServers, conf)
	if err != nil {
		return errors.Wrap(err, "new kafka client error")
	}

	w.producer, err = sarama.NewSyncProducerFromClient(w.client)
	//w.producer, err = sarama.NewAsyncProducerFromClient(w.client)
	if err != nil {
		return errors.Wrap(err, "new kafka producer error")
	}
	return nil
}

func (w *KafkaWriter) Type() string {
	return "kafka"
}

func (w *KafkaWriter) Write(ctx map[string]interface{}) error {
	b, err := json.Marshal(ctx)
	if err != nil {
		return errors.Wrap(err, "json marshal error")
	}

	_, _, err = w.producer.SendMessage(&sarama.ProducerMessage{
		Topic: w.Topic,
		Key:   nil,
		Value: sarama.ByteEncoder(b),
	})
	if err != nil {
		return errors.Wrap(err, "kafka send error")
	}
	return nil
}

func NewKafkaWriter(config yaml.Node) (*KafkaWriter, error) {
	writer := &KafkaWriter{}

	err := writer.init(config)
	if err != nil {
		return nil, errors.Wrap(err, "kafka writer init error")
	}

	return writer, nil
}
