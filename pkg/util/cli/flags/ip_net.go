package flags

import (
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

	name    string
	flagSet *pflag.FlagSet
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

func (f *IPNet) Parse() func() error {
	return nil
}

func (f *IPNet) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *IPNet) AddToFlagSet(name string, flags *pflag.FlagSet) {
	f.name = name
	f.flagSet = flags

	flags.IPNetVarP(f.Target, name, f.Short, f.Default, f.Usage)
	if f.Required {
		markFlagRequired(name, flags)
	}
}
