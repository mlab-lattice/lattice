package flags

import (
	"github.com/spf13/pflag"
)

type Int32 struct {
	Required bool
	Default  int32
	Short    string
	Usage    string
	Target   *int32

	name    string
	flagSet *pflag.FlagSet
}

func (f *Int32) IsRequired() bool {
	return f.Required
}

func (f *Int32) GetShort() string {
	return f.Short
}

func (f *Int32) GetUsage() string {
	return f.Usage
}

func (f *Int32) Parse() func() error {
	return nil
}

func (f *Int32) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *Int32) AddToFlagSet(name string, flags *pflag.FlagSet) {
	f.name = name
	f.flagSet = flags

	flags.Int32VarP(f.Target, name, f.Short, f.Default, f.Usage)
	if f.Required {
		markFlagRequired(name, flags)
	}
}
