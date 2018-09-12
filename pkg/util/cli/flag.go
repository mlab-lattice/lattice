package cli

import (
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
