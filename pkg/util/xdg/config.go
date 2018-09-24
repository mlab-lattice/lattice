package xdg

import (
	"os"
	"path/filepath"
)

const (
	ConfigHomeEnv = "XDG_CONFIG_HOME"
)

func ConfigDir(name string) string {
	configDir := os.Getenv(ConfigHomeEnv)
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}

	return filepath.Join(configDir, name)
}
