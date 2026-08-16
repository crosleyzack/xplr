package tui

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/Netflix/go-env"
	"github.com/crosleyzack/wndr/pkg/keys"
	"github.com/crosleyzack/wndr/pkg/modules/tree"
	"github.com/crosleyzack/wndr/pkg/styles"
)

type configLoc struct {
	FileLoc   string `env:"WNDR_CONFIG"`
	ConfigDir string `env:"XDG_CONFIG_HOME"`
}

type Config struct {
	styles.StyleConfig
	keys.KeyConfig
	tree.TreeConfig
}

const (
	configPath = "/wndr/config.toml"
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
			slog.Warn(fmt.Sprintf("path doesn't exist %s\n", path))
			return &c, nil
		}
		slog.Error(fmt.Sprintf("failed to stat %s\n", path))
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to read %s\n", path))
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return &c, nil
}
