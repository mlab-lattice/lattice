package flags

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/spf13/pflag"
)

type Path struct {
	Required bool
	Default  tree.Path
	Short    string
	Usage    string
	Target   *tree.Path

	target  *string
	name    string
	flagSet *pflag.FlagSet
}

func (f *Path) IsRequired() bool {
	return f.Required
}

func (f *Path) GetShort() string {
	return f.Short
}

func (f *Path) GetUsage() string {
	return f.Usage
}

func (f *Path) Parse() func() error {
	return f.parse
}

func (f *Path) parse() error {
	// if the flag wasn't set and the default wasn't set, then this means that the flag
	// is not required and doesn't have a default. assume the user will check that the
	// flag was not set before using the value
	if !f.Set() && f.Default == "" {
		return nil
	}

	p, err := tree.NewPath(*f.target)
	if err != nil {
		return err
	}

	*f.Target = p
	return nil
}

func (f *Path) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *Path) AddToFlagSet(name string, flags *pflag.FlagSet) {
	t := new(string)
	f.target = t
	flags.StringVarP(f.target, name, f.Short, f.Default.String(), f.Usage)

	if f.Required {
		markFlagRequired(name, flags)
	}
}
