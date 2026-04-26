package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level portwatch configuration.
type Config struct {
	ScanInterval time.Duration `yaml:"scan_interval"`
	Ports        []PortConfig  `yaml:"ports"`
	Alerts       AlertConfig   `yaml:"alerts"`
}

// PortConfig describes a single port/service to monitor.
type PortConfig struct {
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"` // tcp or udp
	Name     string `yaml:"name"`
}

// AlertConfig holds notification settings.
type AlertConfig struct {
	WebhookURL string      `yaml:"webhook_url"`
	Email      EmailConfig `yaml:"email"`
}

// EmailConfig holds SMTP settings for email alerts.
type EmailConfig struct {
	SMTPHost string   `yaml:"smtp_host"`
	SMTPPort int      `yaml:"smtp_port"`
	From     string   `yaml:"from"`
	To       []string `yaml:"to"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validation: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.ScanInterval <= 0 {
		c.ScanInterval = 30 * time.Second
	}
	for i, p := range c.Ports {
		if p.Port < 1 || p.Port > 65535 {
			return fmt.Errorf("port[%d]: invalid port number %d", i, p.Port)
		}
		if p.Protocol == "" {
			c.Ports[i].Protocol = "tcp"
		}
		if p.Protocol != "tcp" && p.Protocol != "udp" {
			return fmt.Errorf("port[%d]: protocol must be tcp or udp, got %q", i, p.Protocol)
		}
	}
	return nil
}
