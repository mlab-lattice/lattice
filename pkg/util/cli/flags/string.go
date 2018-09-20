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

type StringSlice struct {
	Required bool
	Default  []string
	Short    string
	Usage    string
	Target   *[]string

	name    string
	flagSet *pflag.FlagSet
}

func (f *StringSlice) IsRequired() bool {
	return f.Required
}

func (f *StringSlice) GetShort() string {
	return f.Short
}

func (f *StringSlice) GetUsage() string {
	return f.Usage
}

func (f *StringSlice) Parse() func() error {
	return nil
}

func (f *StringSlice) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *StringSlice) AddToFlagSet(name string, flags *pflag.FlagSet) {
	f.name = name
	f.flagSet = flags

	flags.StringSliceVarP(f.Target, name, f.Short, f.Default, f.Usage)
	if f.Required {
		markFlagRequired(name, flags)
	}
}

type StringArray struct {
	Required bool
	Default  string
	Short    string
	Usage    string
	Target   *[]string

	name    string
	flagSet *pflag.FlagSet
}

func (f *StringArray) IsRequired() bool {
	return f.Required
}

func (f *StringArray) GetShort() string {
	return f.Short
}

func (f *StringArray) GetUsage() string {
	return f.Usage
}

func (f *StringArray) Parse() func() error {
	return nil
}

func (f *StringArray) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *StringArray) AddToFlagSet(name string, flags *pflag.FlagSet) {
	f.name = name
	f.flagSet = flags

	flags.StringArrayVarP(f.Target, name, f.Short, nil, f.Usage)
	if f.Required {
		markFlagRequired(name, flags)
	}
}
