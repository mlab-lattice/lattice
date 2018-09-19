package flags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
)

func NewFlagsNotSetError(flags []string) error {
	// this is intended to match https://github.com/spf13/cobra/blob/8d114be902bc9f08717804830a55c48378108a28/command.go#L897
	return fmt.Errorf(`required flag(s) "%s" not set`, strings.Join(flags, `", "`))
}

func markFlagRequired(name string, flags *pflag.FlagSet) {
	cobra.MarkFlagRequired(flags, name)
}
