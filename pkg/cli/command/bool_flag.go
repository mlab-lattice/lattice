package command

import (
	"fmt"

	"github.com/spf13/pflag"
)

type BoolFlag struct {
	Name     string
	Required bool
	Default  bool
	Short    string
	Usage    string
	Target   *bool
}

func (f *BoolFlag) GetName() string {
	return f.Name
}

func (f *BoolFlag) IsRequired() bool {
	return f.Required
}

func (f *BoolFlag) GetShort() string {
	return f.Short
}

func (f *BoolFlag) GetUsage() string {
	return f.Usage
}

func (f *BoolFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *BoolFlag) GetTarget() interface{} {
	return f.Target
}

func (f *BoolFlag) Parse() func() error {
	return nil
}

func (f *BoolFlag) AddToFlagSet(flags *pflag.FlagSet) {
	flags.BoolVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
