package writer

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var logger = logrus.StandardLogger()

func SetLogger(l *logrus.Logger) {
	logger = l
}

type commonConfig struct {
	Type string `yaml:"type"`
}

type Writer interface {
	Type() string
	Write(ctx map[string]interface{}) error
}

var writers = make(map[string]Writer)

func Load(config map[string]yaml.Node) error {
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	var w Writer
	var err error
	w, err = NewConsoleWriter()
	if err != nil {
		return errors.Wrap(err, "create default writer error")
	}
	writers["default"] = w
	for name, cfg := range config {
		w, err = newWriter(cfg)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("create writer(%s) error", name))
		}
		writers[name] = w
	}
	return nil
}

func newWriter(config yaml.Node) (Writer, error) {
	c := commonConfig{}
	err := config.Decode(&c)
	if err != nil {
		return nil, errors.Wrap(err, "yaml decode error")
	}
	switch c.Type {
	case "console":
		return NewConsoleWriter()
	case "open-search":
		return NewOpenSearchWriter(config)
	case "kafka":
		return NewKafkaWriter(config)
	}
	return nil, errors.New("unknown type")
}

func Write(name string, ctx map[string]interface{}) error {
	if ctx != nil {
		if writer, ok := writers[name]; ok {
			logger.Infof("writing to %s(%s)", name, writer.Type())
			return writer.Write(ctx)
		}
		return errors.New(fmt.Sprintf("writers(%s) not exists", name))
	}
	return nil
}
