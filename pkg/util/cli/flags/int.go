package flags

import (
	"fmt"

	"github.com/spf13/pflag"
)

type Int struct {
	Name     string
	Required bool
	Default  int
	Short    string
	Usage    string
	Target   *int
}

func (f *Int) GetName() string {
	return f.Name
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

func (f *Int) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *Int) GetTarget() interface{} {
	return f.Target
}

func (f *Int) Parse() func() error {
	return nil
}

func (f *Int) AddToFlagSet(flags *pflag.FlagSet) {
	flags.IntVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
