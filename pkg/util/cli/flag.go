package cli

import (
	"strings"

	"github.com/spf13/pflag"
)

type Flag interface {
	GetName() string
	IsRequired() bool
	GetShort() string
	GetUsage() string
	Validate() error
	GetTarget() interface{}
	Parse() func() error
	AddToFlagSet(fs *pflag.FlagSet)
}

type Flags []Flag

func (f Flags) Len() int {
	return len(f)
}

func (f Flags) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f Flags) Less(i, j int) bool {
	return strings.Compare(f[i].GetName(), f[j].GetName()) == -1
}
