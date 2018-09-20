package flags

import (
	"fmt"

	"github.com/spf13/pflag"
)

type Bool struct {
	Name     string
	Required bool
	Default  bool
	Short    string
	Usage    string
	Target   *bool
}

func (f *Bool) GetName() string {
	return f.Name
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

func (f *Bool) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *Bool) GetTarget() interface{} {
	return f.Target
}

func (f *Bool) Parse() func() error {
	return nil
}

func (f *Bool) AddToFlagSet(flags *pflag.FlagSet) {
	flags.BoolVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
