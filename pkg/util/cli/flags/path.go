package flags

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/spf13/pflag"
)

type Path struct {
	Name     string
	Required bool
	Default  tree.Path
	Short    string
	Usage    string
	Target   *tree.Path
	target   string
}

func (f *Path) GetName() string {
	return f.Name
}

func (f *Path) IsRequired() bool {
	return f.Required
}

func (f *Path) GetShort() string {
	return f.Short
}

func (f *Path) GetUsage() string {
	return f.Usage
}

func (f *Path) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *Path) GetTarget() interface{} {
	return f.Target
}

func (f *Path) Parse() func() error {
	return func() error {
		p, err := tree.NewPath(f.target)
		if err != nil {
			return err
		}

		*f.Target = p
		return nil
	}
}

func (f *Path) AddToFlagSet(flags *pflag.FlagSet) {
	flags.StringVarP(&f.target, f.Name, f.Short, string(f.Default), f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
