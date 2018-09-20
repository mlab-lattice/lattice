package flags

import (
	"github.com/spf13/pflag"
)

type Int struct {
	Required bool
	Default  int
	Short    string
	Usage    string
	Target   *int

	name    string
	flagSet *pflag.FlagSet
}

func (f *Int) IsRequired() bool {
	return f.Required
}

func (f *Int) GetShort() string {
	return f.Short
}

func (f *Int) GetUsage() string {
	return f.Usage
}

func (f *Int) Parse() func() error {
	return nil
}

func (f *Int) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *Int) AddToFlagSet(name string, flags *pflag.FlagSet) {
	f.name = name
	f.flagSet = flags

	flags.IntVarP(f.Target, name, f.Short, f.Default, f.Usage)
	if f.Required {
		markFlagRequired(name, flags)
	}
}
