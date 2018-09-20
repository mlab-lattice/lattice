package flags

import (
	"fmt"

	"github.com/spf13/pflag"
)

type String struct {
	Name     string
	Required bool
	Default  string
	Short    string
	Usage    string
	Target   *string
}

func (f *String) GetName() string {
	return f.Name
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

func (f *String) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *String) GetTarget() interface{} {
	return f.Target
}

func (f *String) Parse() func() error {
	return nil
}

func (f *String) AddToFlagSet(flags *pflag.FlagSet) {
	flags.StringVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}

type StringSliceFlag struct {
	Name     string
	Required bool
	Default  []string
	Short    string
	Usage    string
	Target   *[]string
}

func (f *StringSliceFlag) GetName() string {
	return f.Name
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

func (f *StringSliceFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *StringSliceFlag) GetTarget() interface{} {
	return f.Target
}

func (f *StringSliceFlag) Parse() func() error {
	return nil
}

func (f *StringSliceFlag) AddToFlagSet(flags *pflag.FlagSet) {
	flags.StringSliceVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}

type StringArrayFlag struct {
	Name     string
	Required bool
	Default  string
	Short    string
	Usage    string
	Target   *[]string
}

func (f *StringArrayFlag) GetName() string {
	return f.Name
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

func (f *StringArrayFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *StringArrayFlag) GetTarget() interface{} {
	return f.Target
}

func (f *StringArrayFlag) Parse() func() error {
	return nil
}

func (f *StringArrayFlag) AddToFlagSet(flags *pflag.FlagSet) {
	flags.StringArrayVarP(f.Target, f.Name, f.Short, nil, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
