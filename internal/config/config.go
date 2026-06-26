package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	ACME     ACMEConfig     `yaml:"acme"`
	Storage  StorageConfig  `yaml:"storage"`
	Sched    SchedulerConfig `yaml:"scheduler"`
	Notify   NotifyConfig   `yaml:"notification"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type ACMEConfig struct {
	Directory string `yaml:"directory"`
}

type StorageConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type SchedulerConfig struct {
	CheckInterval   string `yaml:"check_interval"`
	RenewBeforeDays int    `yaml:"renew_before_days"`
}

type NotifyConfig struct {
	Email   string `yaml:"email"`
	Webhook string `yaml:"webhook"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Sched.RenewBeforeDays == 0 {
		cfg.Sched.RenewBeforeDays = 30
	}
	return cfg, nil
}
