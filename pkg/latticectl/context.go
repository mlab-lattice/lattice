package latticectl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type Context interface {
	Lattice() string
	System() v1.SystemID
}

type DefaultContext struct {
	lattice string
	system  v1.SystemID
}

func (c *DefaultContext) Lattice() string {
	return c.lattice
}

func (c *DefaultContext) System() v1.SystemID {
	return c.system
}

type ContextManager interface {
	Get() (Context, error)
	Set(lattice string, system v1.SystemID) error
}

type ConfigFileContext struct {
	Path      string
	config    *Config
	configSet bool
}

func (c *ConfigFileContext) readConfig() (*Config, error) {
	data, err := ioutil.ReadFile(c.Path)
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

func (c *ConfigFileContext) writeConfig(cfg *Config) error {
	data, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal config: %v", err)
	}

	if err := os.MkdirAll(path.Dir(c.Path), 0755); err != nil {
		return fmt.Errorf("unable to make directory: %v", err)
	}

	if err := ioutil.WriteFile(c.Path, data, 0644); err != nil {
		return fmt.Errorf("unable to write config file: %v", err)
	}

	return nil
}

func (c *ConfigFileContext) Get() (Context, error) {
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

	ctx := &DefaultContext{}
	if cfg != nil && cfg.Context != nil {
		ctx.lattice = cfg.Context.Lattice
		ctx.system = cfg.Context.System
	}

	return ctx, nil
}

func (c *ConfigFileContext) Set(lattice string, system v1.SystemID) error {
	// Want to read the freshest version of the config before overwritting it.
	// N.B.: race condition here against setting other things in the config file
	cfg, err := c.readConfig()
	if err != nil {
		return err
	}

	if cfg == nil {
		cfg = &Config{}
	}

	if cfg.Context == nil {
		cfg.Context = &ConfigContext{}
	}

	cfg.Context.Lattice = lattice
	cfg.Context.System = system

	c.config = cfg
	c.configSet = true

	return c.writeConfig(cfg)
}

type DefaultFileContext struct {
	ctx *ConfigFileContext
}

func (c *DefaultFileContext) setPath() error {
	if c.ctx == nil {
		usr, err := user.Current()
		if err != nil {
			return fmt.Errorf("unable to retrieve current user: %v", err)
		}

		c.ctx = &ConfigFileContext{
			Path: fmt.Sprintf("%v/.latticectl/config", usr.HomeDir),
		}
	}

	return nil
}

func (c *DefaultFileContext) Get() (Context, error) {
	if err := c.setPath(); err != nil {
		return nil, err
	}

	return c.ctx.Get()
}

func (c *DefaultFileContext) Set(lattice string, system v1.SystemID) error {
	if err := c.setPath(); err != nil {
		return err
	}

	return c.ctx.Set(lattice, system)
}
