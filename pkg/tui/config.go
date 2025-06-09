package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/Netflix/go-env"
	"github.com/crosleyzack/xplr/pkg/keys"
	"github.com/crosleyzack/xplr/pkg/modules/tree"
	"github.com/crosleyzack/xplr/pkg/styles"
)

type configLoc struct {
	FileLoc   string `env:"XPLR_CONFIG"`
	ConfigDir string `env:"XDG_CONFIG_HOME,default=~/.config"`
}

type Config struct {
	styles.StyleConfig
	keys.KeyConfig
	tree.TreeConfig
}

// NewConfig creates a new Config object
func NewConfig() (*Config, error) {
	// Load service config
	conf := configLoc{}
	if _, err := env.UnmarshalFromEnviron(&conf); err != nil {
		return nil, fmt.Errorf("failed to read config location: %w", err)
	}
	path := conf.FileLoc
	if path == "" {
		path = filepath.Join(conf.ConfigDir, "/xplr/config.toml")
	}
	var c Config
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &c, nil
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return &c, nil
}
