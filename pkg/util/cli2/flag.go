package cli

import (
	"github.com/spf13/pflag"
)

type Flag interface {
	IsRequired() bool
	GetShort() string
	GetUsage() string
	Value() interface{}
	Parse() func() error
	AddToFlagSet(name string, fs *pflag.FlagSet)
}

type Flags map[string]Flag
