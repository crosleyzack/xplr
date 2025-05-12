package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/Netflix/go-env"
	"github.com/crosleyzack/xplr/internal/keys"
	"github.com/crosleyzack/xplr/internal/styles"
)

type configLoc struct {
	FileLoc string `env:"XPLR_CONFIG" envDefault:"$XDG_CONFIG_HOME/xplr/config.toml"`
}

type Config struct {
	styles.StyleConfig
	keys.KeyConfig
}

// NewConfig creates a new Config object
func NewConfig() (*Config, error) {
	// Load service config
	conf := configLoc{}
	if _, err := env.UnmarshalFromEnviron(&conf); err != nil {
		return nil, fmt.Errorf("failed to read config location: %w", err)
	}
	var c Config
	if _, err := os.Stat(conf.FileLoc); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &c, nil
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	_, err := toml.DecodeFile(conf.FileLoc, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return &c, nil
}
