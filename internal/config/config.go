package config

import (
	"os"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	ACME     ACMEConfig     `yaml:"acme"`
	Storage  StorageConfig  `yaml:"storage"`
	Sched    SchedulerConfig `yaml:"scheduler"`
	Notify   NotifyConfig   `yaml:"notification"`
	Auth     AuthConfig     `yaml:"auth"`
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

type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
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

// ApplySystemConfig overrides the YAML-based config with values from the database SystemConfig row.
// Fields with zero / empty values in DB are left unchanged.
func (c *Config) ApplySystemConfig(sc *store.SystemConfig) {
	if sc.ACMEDirectory != "" {
		c.ACME.Directory = sc.ACMEDirectory
	}
	if sc.CheckInterval != "" {
		c.Sched.CheckInterval = sc.CheckInterval
	}
	if sc.RenewBeforeDays > 0 {
		c.Sched.RenewBeforeDays = sc.RenewBeforeDays
	}
	if sc.NotifyEmail != "" {
		c.Notify.Email = sc.NotifyEmail
	}
	if sc.NotifyWebhook != "" {
		c.Notify.Webhook = sc.NotifyWebhook
	}
	if sc.JWTSecret != "" && sc.JWTSecret != "change-me-in-production" {
		c.Auth.JWTSecret = sc.JWTSecret
	}
}
