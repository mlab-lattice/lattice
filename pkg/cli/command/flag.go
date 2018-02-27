package command

import (
	"github.com/spf13/cobra"
)

type Flag interface {
	required() bool
	short() string
	usage() string
	validate() error
	addToCmd(cmd *cobra.Command, name string)
}
