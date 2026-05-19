package collector

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Simulator struct {
	BaseCollector  `yaml:",inline"`
	DataFile       string        `yaml:"data-file,omitempty"`
	Repeat         bool          `yaml:"repeat,omitempty"`
	RepeatInterval time.Duration `yaml:"repeat-interval,omitempty"`
	TimeField      string        `yaml:"time-field,omitempty"`
	TimeFormat     string        `yaml:"time-format,omitempty"`

	wg       *sync.WaitGroup
	file     *os.File
	scanner  *bufio.Scanner
	startAt  time.Time
	canceled chan struct{}
	sync.Mutex
}

func (c *Simulator) init(config yaml.Node) error {
	if c == nil {
		return errors.New("nil pointer")
	}

	err := config.Decode(c)
	if err != nil {
		return errors.Wrap(err, "yaml decode error")
	}

	if c.TimeField == "" {
		return errors.New("time-field must be defined")
	} else if c.TimeFormat == "" {
		return errors.New("time-format must be defined")
	}

	if c.Repeat && c.RepeatInterval == 0 {
		c.RepeatInterval = time.Minute
	}

	c.logger = logger.WithFields(logrus.Fields{
		"data":    c.DataFile,
		"context": c.Context,
	})

	return nil
}

func (c *Simulator) getTime(ctx map[string]interface{}) (time.Time, error) {
	if t := ctx[c.TimeField]; t != nil {
		timeString, ok := t.(string)
		if ok {
			return time.Parse(c.TimeFormat, timeString)
		}
	}
	return time.Time{}, errors.Errorf("time-field '%s' is not string type", c.TimeField)
}

func (c *Simulator) simulate() {
	if c.file != nil {
		c.Lock()
		c.canceled = make(chan struct{}, 1)
		defer func() {
			c.Unlock()
			c.canceled = nil
		}()
		_, err := c.file.Seek(0, io.SeekStart)
		if err != nil {
			c.logger.WithError(err).Errorf("failed to reset file offset")
			c.canceled = nil
			c.Stop()
		}
		c.scanner = bufio.NewScanner(c.file)
		scanned := c.scanner.Scan()
		if scanned {
			startTime := time.Now()
			b := c.scanner.Bytes()
			var ctx map[string]interface{}
			err := json.Unmarshal(b, &ctx)
			if err != nil {
				c.logger.WithError(err).Errorf("invalid data format: %s", string(b))
				return
			}
			startData, err := c.getTime(ctx)
			if err != nil {
				c.logger.WithError(err).Error()
				startData = startTime
			}

			ctx[c.TimeField] = startTime
			c.Write(ctx)

			for c.scanner.Scan() {
				if len(c.canceled) > 0 {
					c.done()
					return
				}

				b = c.scanner.Bytes()
				ctx = map[string]interface{}{}
				err = json.Unmarshal(b, &ctx)
				if err != nil {
					c.logger.WithError(err).Errorf("invalid data format: %s", string(b))
					continue
				}

				ctxTime, err := c.getTime(ctx)
				sinceStart := ctxTime.Sub(startData)
				converted := startTime.Add(sinceStart)
				timeToWait := time.Until(converted)
				var timer *time.Timer
				if err != nil {
					c.logger.WithError(err).Error()
				} else if timeToWait > 100*time.Millisecond {
					timer = time.NewTimer(timeToWait)
				}

				if timer != nil {
					select {
					case <-c.canceled:
						c.done()
						return
					case <-timer.C:
					}
				}

				ctx[c.TimeField] = converted
				c.Write(ctx)
			}

			if err := c.scanner.Err(); err != nil {
				c.logger.WithError(err).Error("Error while reading file")
				scanned = false
			}
		}

		if scanned && c.Repeat {
			go func() {
				select {
				case <-c.canceled:
					c.done()
					return
				case <-time.NewTimer(c.RepeatInterval).C:
					c.simulate()
				}
			}()
		}
	}
}

func (c *Simulator) Start(wg *sync.WaitGroup) error {
	if c.wg == nil && wg != nil && c.file == nil && c.scanner == nil {
		file, err := os.OpenFile(c.DataFile, os.O_RDONLY, 0)
		if err != nil {
			return errors.Wrap(err, "open file error")
		}
		c.file = file

		c.logger.Info("start simulate")
		go c.simulate()

		c.wg = wg
		c.wg.Add(1)
	}
	return nil
}

func (c *Simulator) done() {
	if c.wg != nil && c.file != nil && c.scanner != nil {
		_ = c.file.Close()
		c.file = nil
		c.scanner = nil
		c.wg.Done()
	}
}

func (c *Simulator) Stop() {
	if c.wg != nil && c.file != nil && c.scanner != nil {
		c.logger.Info("stop simulate")
		if c.canceled != nil {
			c.canceled <- struct{}{}
		} else {
			c.done()
		}
	}
}

func NewSimulator(config yaml.Node) (*Simulator, error) {
	simulator := &Simulator{}

	err := simulator.init(config)
	if err != nil {
		return nil, errors.Wrap(err, "simulator init error")
	}

	simulator.logger.Info("create simulator")

	return simulator, nil
}
