package flags

import (
	"github.com/spf13/pflag"
)

type Bool struct {
	Required bool
	Default  bool
	Short    string
	Usage    string
	Target   *bool

	name    string
	flagSet *pflag.FlagSet
}

func (f *Bool) IsRequired() bool {
	return f.Required
}

func (f *Bool) GetShort() string {
	return f.Short
}

func (f *Bool) GetUsage() string {
	return f.Usage
}

func (f *Bool) Parse() func() error {
	return nil
}

func (f *Bool) Value() interface{} {
	return f.Target
}

func (f *Bool) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *Bool) AddToFlagSet(name string, flags *pflag.FlagSet) {
	f.name = name
	f.flagSet = flags

	flags.BoolVarP(f.Target, name, f.Short, f.Default, f.Usage)
	if f.Required {
		markFlagRequired(name, flags)
	}
}
