package collector

import (
	"monitoring_server/pkg/writer"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var logger = logrus.StandardLogger()

func SetLogger(l *logrus.Logger) {
	logger = l
}

type Collector interface {
	Start(wg *sync.WaitGroup) error
	Stop()
}

type Collectors []Collector

func (c Collectors) Start(wg *sync.WaitGroup) error {
	logger.Infof("Start %d collectors", len(c))
	var i int
	var col Collector
	for i, col = range collectors {
		err := col.Start(wg)
		if err != nil {
			if i > 0 {
				c[:i-1].Stop()
			}
			return errors.Wrap(err, "collectors start error")
		}
		<-time.NewTimer(1 * time.Second).C
	}

	return nil
}

func (c Collectors) Stop() {
	logger.Info("Stop collectors")
	for _, col := range c {
		col.Stop()
	}
}

type BaseCollector struct {
	Writers []string          `yaml:"writers,omitempty"`
	Context map[string]string `yaml:"context,omitempty"`

	logger *logrus.Entry
}

func (c *BaseCollector) Write(ctx map[string]interface{}) {
	if c != nil {
		go func() {
			for k, v := range c.Context {
				ctx[k] = v
			}
			for _, w := range c.Writers {
				err := writer.Write(w, ctx)
				if err != nil {
					logger.Errorf("%s\n", err.Error())
				}
			}
		}()
	}
}

type config struct {
	Scheduler    []yaml.Node `yaml:"scheduler,omitempty"`
	SnmpTrap     []yaml.Node `yaml:"snmp-trap,omitempty"`
	HttpListener []yaml.Node `yaml:"http-listener,omitempty"`
	Simulator    []yaml.Node `yaml:"simulator,omitempty"`
}

var collectors Collectors

func Load(cfg yaml.Node) error {
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	c := config{}

	err := cfg.Decode(&c)
	if err != nil {
		return errors.Wrap(err, "yaml decode error")
	}

	logger.Info("Load All Collectors")

	for _, v := range c.Simulator {
		simulator, err := NewSimulator(v)
		if err != nil {
			return errors.Wrap(err, "create simulator error")
		}
		collectors = append(collectors, simulator)
	}

	return nil
}

func StartAll(wg *sync.WaitGroup) error {
	return collectors.Start(wg)
}

func StopAll() {
	collectors.Stop()
}
