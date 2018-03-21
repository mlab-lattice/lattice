package command

import (
	"fmt"

	"github.com/spf13/pflag"
)

type IntFlag struct {
	Name     string
	Required bool
	Default  int
	Short    string
	Usage    string
	Target   *int
}

func (f *IntFlag) GetName() string {
	return f.Name
}

func (f *IntFlag) IsRequired() bool {
	return f.Required
}

func (f *IntFlag) GetShort() string {
	return f.Short
}

func (f *IntFlag) GetUsage() string {
	return f.Usage
}

func (f *IntFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *IntFlag) GetTarget() interface{} {
	return f.Target
}

func (f *IntFlag) Parse() func() error {
	return nil
}

func (f *IntFlag) AddToFlagSet(flags *pflag.FlagSet) {
	flags.IntVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
