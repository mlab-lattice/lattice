package flags

import (
	"fmt"
	"net"

	"github.com/spf13/pflag"
)

type IPNet struct {
	Name     string
	Required bool
	Default  net.IPNet
	Short    string
	Usage    string
	Target   *net.IPNet
}

func (f *IPNet) GetName() string {
	return f.Name
}

func (f *IPNet) IsRequired() bool {
	return f.Required
}

func (f *IPNet) GetShort() string {
	return f.Short
}

func (f *IPNet) GetUsage() string {
	return f.Usage
}

func (f *IPNet) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *IPNet) GetTarget() interface{} {
	return f.Target
}

func (f *IPNet) Parse() func() error {
	return nil
}

func (f *IPNet) AddToFlagSet(flags *pflag.FlagSet) {
	flags.IPNetVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
