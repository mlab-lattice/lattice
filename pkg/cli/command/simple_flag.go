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

func (f *StringFlag) GetName() string {
	return f.Name
}

func (f *StringFlag) IsRequired() bool {
	return f.Required
}

func (f *StringFlag) GetShort() string {
	return f.Short
}

func (f *StringFlag) GetUsage() string {
	return f.Usage
}

func (f *StringFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *StringFlag) AddToCmd(cmd *cobra.Command) {
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

func (f *IntFlag) GetName() string {
	return f.Name
}

func (f *IntFlag) IsRequired() bool {
	return f.Required
}

func (f *IntFlag) GetShort() string {
	return f.Short
}

func (f *IntFlag) GetUsage() string {
	return f.Usage
}

func (f *IntFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *IntFlag) AddToCmd(cmd *cobra.Command) {
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

func (f *BoolFlag) GetName() string {
	return f.Name
}

func (f *BoolFlag) IsRequired() bool {
	return f.Required
}

func (f *BoolFlag) GetShort() string {
	return f.Short
}

func (f *BoolFlag) GetUsage() string {
	return f.Usage
}

func (f *BoolFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *BoolFlag) AddToCmd(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		cmd.MarkFlagRequired(f.Name)
	}
}
