package main

import (
	"gopkg.in/yaml.v2"
)

type ConfigItem struct {
	Idekey  string `yaml:"idekey"`
	Address string `yaml:"address"`
}

type Config struct {
	XdebugAddr   string       `yaml:"xdebug-address"`
	RegistryAddr string       `yaml:"registry-address"`
	Predefined   []ConfigItem `yaml:"predefined,flow"`
}

func parseConfig(data []byte) (*Config, error) {
	config := &Config{}
	err := yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
