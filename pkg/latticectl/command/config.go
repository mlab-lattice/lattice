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

// ConfigFile is a configuration file holding latticectl configuration.
type ConfigFile struct {
	Path string

	config       Config
	configLoaded bool
}

// Config is latticectl configuration.
type Config struct {
	CurrentContext string             `json:"currentContext"`
	Contexts       map[string]Context `json:"contexts"`
}

// Contexts returns the Contexts defined in the ConfigFile's Config.
func (c *ConfigFile) Contexts() (map[string]Context, error) {
	if err := c.load(); err != nil {
		return nil, err
	}

	return c.config.Contexts, nil
}

// Contexts returns the name of the Context currently selected in the ConfigFile's Config.
func (c *ConfigFile) CurrentContext() (string, error) {
	if err := c.load(); err != nil {
		return "", err
	}

	return c.config.CurrentContext, nil
}

// Context returns the Context for a given name in the ConfigFile's Config.
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

// CreateContext adds a new Context with the given name to the ConfigFile's Config and flushes the change.
func (c *ConfigFile) CreateContext(name string, context *Context) error {
	cfg, err := c.Config()
	if err != nil {
		return err
	}

	c.UpdateContext(name, context)
	cfg.CurrentContext = name
	return c.save()
}

// DeleteContext removes the Context with the given name from the ConfigFile's Config and flushes the change.
func (c *ConfigFile) DeleteContext(name string) error {
	cfg, err := c.Config()
	if err != nil {
		return err
	}

	delete(cfg.Contexts, name)
	return c.save()
}

// UpdateContext replaces the Context with the given name with the supplied Context in the ConfigFile's Config
// and flushes the change.
func (c *ConfigFile) UpdateContext(name string, context *Context) error {
	cfg, err := c.Config()
	if err != nil {
		return err
	}

	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]Context)
	}

	cfg.Contexts[name] = *context
	return c.save()
}

// SetCurrentContext changes the currently selected Context in the ConfigFile's Config to the one
// with the given name and flushes the change.
func (c *ConfigFile) SetCurrentContext(name string) error {
	cfg, err := c.Config()
	if err != nil {
		return err
	}

	if _, ok := cfg.Contexts[name]; !ok {
		return NewInvalidContextError(name)
	}

	cfg.CurrentContext = name
	return c.save()
}

// UnsetCurrentContext unselects the currently selected Context in the ConfigFile's Config and flushes the change.
func (c *ConfigFile) UnsetCurrentContext() error {
	cfg, err := c.Config()
	if err != nil {
		return err
	}

	cfg.CurrentContext = ""
	return c.save()
}

// Config loads the ConfigFile's Config and returns it.
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
