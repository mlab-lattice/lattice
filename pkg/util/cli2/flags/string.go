package flags

import (
	"github.com/spf13/pflag"
)

type String struct {
	Required bool
	Default  string
	Short    string
	Usage    string
	Target   *string

	name    string
	flagSet *pflag.FlagSet
}

func (f *String) IsRequired() bool {
	return f.Required
}

func (f *String) GetShort() string {
	return f.Short
}

func (f *String) GetUsage() string {
	return f.Usage
}

func (f *String) Parse() func() error {
	return nil
}

func (f *String) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *String) AddToFlagSet(name string, flags *pflag.FlagSet) {
	f.name = name
	f.flagSet = flags

	flags.StringVarP(f.Target, name, f.Short, f.Default, f.Usage)
	if f.Required {
		markFlagRequired(name, flags)
	}
}

type StringSliceFlag struct {
	Required bool
	Default  []string
	Short    string
	Usage    string
	Target   *[]string

	name    string
	flagSet *pflag.FlagSet
}

func (f *StringSliceFlag) IsRequired() bool {
	return f.Required
}

func (f *StringSliceFlag) GetShort() string {
	return f.Short
}

func (f *StringSliceFlag) GetUsage() string {
	return f.Usage
}

func (f *StringSliceFlag) Parse() func() error {
	return nil
}

func (f *StringSliceFlag) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *StringSliceFlag) AddToFlagSet(name string, flags *pflag.FlagSet) {
	f.name = name
	f.flagSet = flags

	flags.StringSliceVarP(f.Target, name, f.Short, f.Default, f.Usage)
	if f.Required {
		markFlagRequired(name, flags)
	}
}

type StringArrayFlag struct {
	Required bool
	Default  string
	Short    string
	Usage    string
	Target   *[]string

	name    string
	flagSet *pflag.FlagSet
}

func (f *StringArrayFlag) IsRequired() bool {
	return f.Required
}

func (f *StringArrayFlag) GetShort() string {
	return f.Short
}

func (f *StringArrayFlag) GetUsage() string {
	return f.Usage
}

func (f *StringArrayFlag) Parse() func() error {
	return nil
}

func (f *StringArrayFlag) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *StringArrayFlag) AddToFlagSet(name string, flags *pflag.FlagSet) {
	f.name = name
	f.flagSet = flags

	flags.StringArrayVarP(f.Target, name, f.Short, nil, f.Usage)
	if f.Required {
		markFlagRequired(name, flags)
	}
}
