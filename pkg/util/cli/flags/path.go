package flags

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/spf13/pflag"
)

type PathFlag struct {
	Name     string
	Required bool
	Default  tree.Path
	Short    string
	Usage    string
	Target   *tree.Path
}

func (f *PathFlag) GetName() string {
	return f.Name
}

func (f *PathFlag) IsRequired() bool {
	return f.Required
}

func (f *PathFlag) GetShort() string {
	return f.Short
}

func (f *PathFlag) GetUsage() string {
	return f.Usage
}

func (f *PathFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *PathFlag) GetTarget() interface{} {
	return f.Target
}

func (f *PathFlag) Parse() func() error {
	return nil
}

func (f *PathFlag) AddToFlagSet(flags *pflag.FlagSet) {
	//flags.StringVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)
	//
	//if f.Required {
	//	markFlagRequired(f.Name, flags)
	//}
}
