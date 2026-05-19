package main

import (
	"monitoring_server/pkg/common/log"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Writer    map[string]yaml.Node `yaml:"writer,omitempty"`
	Collector yaml.Node            `yaml:"collector,omitempty"`
}

var defaultConfig = &Config{}

func loadYamlFile(in string, out interface{}) error {
	log.GetLogger().Infof("Read config file: %s", in)
	b, err := os.ReadFile(in)
	if err != nil {
		return errors.Wrap(err, "read file error")
	}

	err = yaml.Unmarshal(b, out)
	if err != nil {
		return errors.Wrap(err, "yaml unmarshal error")
	}

	return nil
}
