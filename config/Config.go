package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppName    string `envconfig:"APP_NAME" default:"rsclabs-pavel"`
	AppVersion string `envconfig:"APP_VERSION" default:"1.0.0"`
	MaxBanners int    `envconfig:"MAX_BANNERS" default:"100"`
	Port       string `envconfig:"PORT" default:"8080"`
}

func NewConfig() *Config {
	var cnf Config
	if err := envconfig.Process("", &cnf); err != nil {
		panic(fmt.Errorf("error environmaent variable parsing: %w", err))
	}
	return &cnf
}
