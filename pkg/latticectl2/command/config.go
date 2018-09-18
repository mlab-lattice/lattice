package command

import (
	"encoding/json"
	"fmt"
	"github.com/mlab-lattice/lattice/pkg/util/xdg"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

const (
	Latticectl        = "latticectl"
	DefaultConfigFile = "config.json"
)

type ConfigFile struct {
	Path string

	config       Config
	configLoaded bool
}

type Config struct {
	CurrentContext string             `json:"currentContext"`
	Contexts       map[string]Context `json:"contexts"`
}

func (c *ConfigFile) Contexts() (map[string]Context, error) {
	if err := c.load(); err != nil {
		return nil, err
	}

	return c.config.Contexts, nil
}

func (c *ConfigFile) CurrentContext() (string, error) {
	if err := c.load(); err != nil {
		return "", err
	}

	return c.config.CurrentContext, nil
}

func (c *ConfigFile) Context(name string) (*Context, error) {
	if name == "" {
		return nil, NewNoContextSetError()
	}

	if err := c.load(); err != nil {
		return nil, err
	}

	ctx, ok := c.config.Contexts[name]
	if !ok {
		return nil, NewInvalidContextError(name)
	}

	return &ctx, nil
}

func (c *ConfigFile) CreateContext(name string, context Context) error {
	cfg, err := c.Config()
	if err != nil {
		return err
	}

	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]Context)
	}

	cfg.Contexts[name] = context
	cfg.CurrentContext = name
	return c.save()
}

func (c *ConfigFile) SetCurrentContext(context string) error {
	// Want to read the freshest version of the config before overwritting it.
	// N.B.: race condition here against setting other things in the config file
	cfg, err := c.Config()
	if err != nil {
		return err
	}

	if _, ok := cfg.Contexts[context]; !ok {
		return NewInvalidContextError(context)
	}

	cfg.CurrentContext = context
	return c.save()
}

func (c *ConfigFile) Config() (*Config, error) {
	if err := c.load(); err != nil {
		return nil, err
	}

	return &c.config, nil
}

func (c *ConfigFile) load() error {
	if c.configLoaded {
		return nil
	}

	cfg := Config{}
	data, err := ioutil.ReadFile(c.configFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			c.config = cfg
			return nil
		}

		return fmt.Errorf("unable to read config file: %v", err)
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return fmt.Errorf("error parsing config file: %v", err)
	}

	c.config = cfg
	c.configLoaded = true
	return nil
}

func (c *ConfigFile) save() error {
	if err := c.writeConfig(c.config); err != nil {
		return err
	}

	c.configLoaded = true
	return nil
}

func (c *ConfigFile) readConfig() (Config, error) {
	cfg := Config{}
	data, err := ioutil.ReadFile(c.configFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}

		return cfg, fmt.Errorf("unable to read config file: %v", err)
	}

	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

func (c *ConfigFile) writeConfig(cfg Config) error {
	data, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal config: %v", err)
	}

	file := c.configFilePath()
	if err := os.MkdirAll(path.Dir(file), 0755); err != nil {
		return fmt.Errorf("unable to make directory: %v", err)
	}

	if err := ioutil.WriteFile(file, data, 0644); err != nil {
		return fmt.Errorf("unable to write config file: %v", err)
	}

	return nil
}

func (c *ConfigFile) configFilePath() string {
	if c.Path != "" {
		return c.Path
	}

	return filepath.Join(xdg.ConfigDir(Latticectl), DefaultConfigFile)
}
