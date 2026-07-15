package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	ACME    ACMEConfig    `yaml:"acme"`
	Storage StorageConfig `yaml:"storage"`
	Auth    AuthConfig    `yaml:"auth"`
}

type ServerConfig struct {
	Port      int    `yaml:"port"`
	ProxyURL  string `yaml:"proxy_url"`
	DNSListen string `yaml:"dns_listen"`
}

type ACMEConfig struct {
	Directory             string   `yaml:"directory"`
	RecursiveNameservers  []string `yaml:"recursive_nameservers"`
}

type StorageConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
	DeployKey string `yaml:"deploy_key"`
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
	if cfg.ACME.Directory == "" {
		cfg.ACME.Directory = "https://acme-v02.api.letsencrypt.org/directory"
	}
	if len(cfg.ACME.RecursiveNameservers) == 0 {
		cfg.ACME.RecursiveNameservers = []string{
			"119.29.29.29:53",
			"223.5.5.5:53",
			"114.114.114.114:53",
			"8.8.8.8:53",
		}
	}
	if cfg.Auth.DeployKey == "" {
		b := make([]byte, 16)
		rand.Read(b)
		cfg.Auth.DeployKey = hex.EncodeToString(b)
		newData, _ := yaml.Marshal(cfg)
		os.WriteFile(path, newData, 0644)
	}
	return cfg, nil
}
