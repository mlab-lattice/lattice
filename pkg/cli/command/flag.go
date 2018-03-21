package command

import (
	"github.com/spf13/cobra"
)

type Flag interface {
	GetName() string
	IsRequired() bool
	GetShort() string
	GetUsage() string
	Validate() error
	AddToCmd(cmd *cobra.Command)
}

type Flags []Flag
