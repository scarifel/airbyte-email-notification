package config

import "github.com/kelseyhightower/envconfig"

type SMTP struct {
	Host            string   `envconfig:"SMTP_HOST" required:"true"`
	Port            int      `envconfig:"SMTP_PORT" required:"true"`
	AnonymousAccess bool     `envconfig:"SMTP_ANONYMOUS_ACCESS" required:"true"`
	TLS             bool     `envconfig:"SMTP_TLS_ENABLE"`
	Username        string   `envconfig:"SMTP_USERNAME"`
	Password        string   `envconfig:"SMTP_PASSWORD"`
	From            string   `envconfig:"SMTP_FROM" required:"true"`
	To              []string `envconfig:"SMTP_TO" required:"true"`
	Subject         string   `envconfig:"SMTP_SUBJECT" default:"AIRBYTE NOTIFICATION"`
}

type App struct {
	Host string `envconfig:"HOST" default:"localhost"`
	Port string `envconfig:"PORT" default:"8080"`
}

type Config struct {
	App  App
	SMTP SMTP
}

func LoadConfig() (*Config, error) {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		return nil, err
	}

	return &config, nil
}
