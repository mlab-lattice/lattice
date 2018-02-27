package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Flag interface {
	required() bool
	short() string
	usage() string
	validate() error
	addToCmd(cmd *cobra.Command, name string)
}

type StringFlag struct {
	Required bool
	Default  string
	Short    string
	Usage    string
	Target   *string
}

func (f *StringFlag) required() bool {
	return f.Required
}

func (f *StringFlag) short() string {
	return f.Short
}

func (f *StringFlag) usage() string {
	return f.Usage
}

func (f *StringFlag) validate() error {
	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *StringFlag) addToCmd(cmd *cobra.Command, name string) {
	cmd.Flags().StringVarP(f.Target, name, f.Short, f.Default, f.Usage)

	if f.Required {
		cmd.MarkFlagRequired(name)
	}
}

type IntFlag struct {
	Required bool
	Default  int
	Short    string
	Usage    string
	Target   *int
}

func (f *IntFlag) required() bool {
	return f.Required
}

func (f *IntFlag) short() string {
	return f.Short
}

func (f *IntFlag) usage() string {
	return f.Usage
}

func (f *IntFlag) validate() error {
	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *IntFlag) addToCmd(cmd *cobra.Command, name string) {
	cmd.Flags().IntVarP(f.Target, name, f.Short, f.Default, f.Usage)

	if f.Required {
		cmd.MarkFlagRequired(name)
	}
}
