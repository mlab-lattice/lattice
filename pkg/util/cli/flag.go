package cli

import (
	"github.com/spf13/pflag"
)

type Flag interface {
	IsRequired() bool
	GetShort() string
	GetUsage() string
	Parse() func() error
	Set() bool
	AddToFlagSet(name string, fs *pflag.FlagSet)
}

type Flags map[string]Flag
