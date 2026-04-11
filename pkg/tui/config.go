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
	ConfigDir string `env:"XDG_CONFIG_HOME"`
}

type Config struct {
	styles.StyleConfig
	keys.KeyConfig
	tree.TreeConfig
}

const (
	configPath = "/xplr/config.toml"
)

// NewConfig creates a new Config object
func NewConfig() (*Config, error) {
	// Load service config
	conf := configLoc{}
	if _, err := env.UnmarshalFromEnviron(&conf); err != nil {
		return nil, fmt.Errorf("failed to read config location: %w", err)
	}
	var path string
	switch {
	case conf.FileLoc != "":
		path = conf.FileLoc
	case conf.ConfigDir != "":
		path = filepath.Join(conf.ConfigDir, configPath)
	default:
		home, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home: %w", err)
		}
		path = filepath.Join(home, configPath)
	}
	var c Config
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("path doesn't exist %s\n", path)
			return &c, nil
		}
		fmt.Printf("failed to stat %s\n", path)
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		fmt.Printf("failed to read %s\n", path)
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return &c, nil
}
