package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func markFlagRequired(name string, flags *pflag.FlagSet) {
	cobra.MarkFlagRequired(flags, name)
}
