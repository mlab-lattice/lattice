package command

import (
	"fmt"

	"github.com/spf13/pflag"
)

type Int32Flag struct {
	Name     string
	Required bool
	Default  int32
	Short    string
	Usage    string
	Target   *int32
}

func (f *Int32Flag) GetName() string {
	return f.Name
}

func (f *Int32Flag) IsRequired() bool {
	return f.Required
}

func (f *Int32Flag) GetShort() string {
	return f.Short
}

func (f *Int32Flag) GetUsage() string {
	return f.Usage
}

func (f *Int32Flag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *Int32Flag) GetTarget() interface{} {
	return f.Target
}

func (f *Int32Flag) Parse() func() error {
	return nil
}

func (f *Int32Flag) AddToFlagSet(flags *pflag.FlagSet) {
	flags.Int32VarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
