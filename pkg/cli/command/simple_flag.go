package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

type StringFlag struct {
	Name     string
	Required bool
	Default  string
	Short    string
	Usage    string
	Target   *string
}

func (f *StringFlag) name() string {
	return f.Name
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
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *StringFlag) addToCmd(cmd *cobra.Command) {
	cmd.Flags().StringVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		cmd.MarkFlagRequired(f.Name)
	}
}

type IntFlag struct {
	Name     string
	Required bool
	Default  int
	Short    string
	Usage    string
	Target   *int
}

func (f *IntFlag) name() string {
	return f.Name
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
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *IntFlag) addToCmd(cmd *cobra.Command) {
	cmd.Flags().IntVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		cmd.MarkFlagRequired(f.Name)
	}
}

type BoolFlag struct {
	Name     string
	Required bool
	Default  bool
	Short    string
	Usage    string
	Target   *bool
}

func (f *BoolFlag) name() string {
	return f.Name
}

func (f *BoolFlag) required() bool {
	return f.Required
}

func (f *BoolFlag) short() string {
	return f.Short
}

func (f *BoolFlag) usage() string {
	return f.Usage
}

func (f *BoolFlag) validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *BoolFlag) addToCmd(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		cmd.MarkFlagRequired(f.Name)
	}
}
