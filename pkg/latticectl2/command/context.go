package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"path/filepath"
)

const (
	configHomeEnv  = "XDG_CONFIG_HOME"
	latticectlName = "latticectl"
	configFileName = "config.json"
)

type Context struct {
	Lattice string      `json:"lattice"`
	System  v1.SystemID `json:"system"`
}

type configFile struct {
	Path      string
	config    *Config
	configSet bool
}

func (c *configFile) GetContext() (*Context, error) {
	// Want to ensure a consistent state of the config
	// throughout a command, so once we read it in
	// we don't want to read it again
	cfg := c.config
	if !c.configSet {
		var err error

		cfg, err = c.readConfig()
		if err != nil {
			return nil, err
		}

		c.config = cfg
		c.configSet = true
	}

	return c.config.Context, nil
}

func (c *configFile) SetContext(ctx *Context) error {
	// Want to read the freshest version of the config before overwritting it.
	// N.B.: race condition here against setting other things in the config file
	cfg, err := c.readConfig()
	if err != nil {
		return err
	}

	if cfg == nil {
		cfg = &Config{}
	}

	cfg.Context = ctx

	c.config = cfg
	c.configSet = true

	return c.writeConfig(cfg)
}

func (c *configFile) readConfig() (*Config, error) {
	data, err := ioutil.ReadFile(c.configFilepath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("unable to read config file: %v", err)
	}

	cfg := Config{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config file: %v", err)
	}

	return &cfg, nil
}

func (c *configFile) writeConfig(cfg *Config) error {
	data, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal config: %v", err)
	}

	if err := os.MkdirAll(path.Dir(c.configFilepath()), 0755); err != nil {
		return fmt.Errorf("unable to make directory: %v", err)
	}

	if err := ioutil.WriteFile(c.Path, data, 0644); err != nil {
		return fmt.Errorf("unable to write config file: %v", err)
	}

	return nil
}

func (c *configFile) configFilepath() string {
	if c.Path != "" {
		return c.Path
	}

	home := os.Getenv(configHomeEnv)
	if home == "" {
		home = os.Getenv("HOME")
	}

	return filepath.Join(home, latticectlName, configFileName)
}
