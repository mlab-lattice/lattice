package cli

import (
	"fmt"
	"net"

	"github.com/spf13/pflag"
)

type IPNetFlag struct {
	Name     string
	Required bool
	Default  net.IPNet
	Short    string
	Usage    string
	Target   *net.IPNet
}

func (f *IPNetFlag) GetName() string {
	return f.Name
}

func (f *IPNetFlag) IsRequired() bool {
	return f.Required
}

func (f *IPNetFlag) GetShort() string {
	return f.Short
}

func (f *IPNetFlag) GetUsage() string {
	return f.Usage
}

func (f *IPNetFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *IPNetFlag) GetTarget() interface{} {
	return f.Target
}

func (f *IPNetFlag) Parse() func() error {
	return nil
}

func (f *IPNetFlag) AddToFlagSet(flags *pflag.FlagSet) {
	flags.IPNetVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
