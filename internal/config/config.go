package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

// Парсинг конфига

type Config struct {
	ListenAddr string   `yaml:"listen_addr"`
	Backends   []string `yaml:"backends"`
	RateLimit  struct {
		Capacity   uint64 `yaml:"capacity"`
		RefillRate uint64 `yaml:"refill_rate"`
	} `yaml:"rate_limit"`
	HealthCheck struct {
		Path        string `yaml:"path"`
		IntervalSec uint64 `yaml:"interval_sec"`
	} `yaml:"health_check"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
