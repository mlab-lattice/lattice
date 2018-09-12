package flags

import (
	"github.com/spf13/pflag"
)

type Bool struct {
	Required bool
	Default  bool
	Short    string
	Usage    string
	target   bool
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
	return nil
}

func (f *Bool) Value() interface{} {
	return f.target
}

func (f *Bool) Parse() func() error {
	return nil
}

func (f *Bool) AddToFlagSet(name string, flags *pflag.FlagSet) {
	flags.BoolVarP(&f.target, name, f.Short, f.Default, f.Usage)

	if f.Required {
		markFlagRequired(name, flags)
	}
}
