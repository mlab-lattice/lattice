package flags

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/util/cli2"

	"github.com/spf13/pflag"
)

// Embedded is a Flag that allows you to parse multiple values from the same flag.
// For example, if c is an embedded flag with two string flags, bar and buzz,
// you could say --c "bar=hello,buzz=world".
type Embedded struct {
	Required bool
	Short    string
	Usage    string
	Flags    cli.Flags

	target  []string
	name    string
	flagSet *pflag.FlagSet
}

func (f *Embedded) IsRequired() bool {
	return f.Required
}

func (f *Embedded) GetShort() string {
	return f.Short
}

func (f *Embedded) GetUsage() string {
	return f.Usage
}

func (f *Embedded) Parse() func() error {
	return f.parse
}

func (f *Embedded) parse() error {
	flags := &pflag.FlagSet{}
	for name, flag := range f.Flags {
		flag.AddToFlagSet(name, flags)
	}

	var dashedValues []string
	for _, value := range f.target {
		dashedValues = append(dashedValues, fmt.Sprintf("--%v", value))
	}

	err := flags.Parse(dashedValues)
	if err != nil {
		return err
	}

	for name, flag := range f.Flags {
		if flag.IsRequired() && !flags.Changed(name) {
			return NewFlagsNotSetError([]string{name})
		}

		parser := flag.Parse()
		if parser != nil {
			err := parser()
			if err != nil {
				return fmt.Errorf("error parsing embedded flag %v: %v", name, err)
			}
		}
	}

	return nil
}

func (f *Embedded) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *Embedded) AddToFlagSet(name string, flags *pflag.FlagSet) {
	f.name = name
	f.flagSet = flags

	flags.StringArrayVar(&f.target, name, nil, f.Usage)
	if f.Required {
		markFlagRequired(name, flags)
	}
}

type DelayedEmbedded struct {
	Name        string
	Required    bool
	Short       string
	Usage       string
	Flags       map[string]cli.Flags
	Delimiter   string
	FlagChooser func() (*string, error)

	target  []string
	name    string
	flagSet *pflag.FlagSet
}

func (f *DelayedEmbedded) IsRequired() bool {
	return f.Required
}

func (f *DelayedEmbedded) GetShort() string {
	return f.Short
}

func (f *DelayedEmbedded) GetUsage() string {
	return f.Usage
}

func (f *DelayedEmbedded) Parse() func() error {
	return f.parse
}

func (f *DelayedEmbedded) parse() error {
	choice, err := f.FlagChooser()
	if err != nil {
		return err
	}

	if choice == nil {
		if f.Required {
			// TODO: pretty obtuse error for a user to receive
			return fmt.Errorf("flag is required, but no choice was made")
		}
		return nil
	}

	flags, ok := f.Flags[*choice]
	if !ok {
		// TODO: pretty obtuse error for a user to receive
		return fmt.Errorf("invalid flag choice %v", choice)
	}

	embedded := &Embedded{
		Required: f.Required,
		Short:    f.Short,
		Usage:    f.Usage,
		Flags:    flags,

		target: f.target,
	}

	return embedded.parse()
}

func (f *DelayedEmbedded) Set() bool {
	return f.flagSet.Changed(f.name)
}

func (f *DelayedEmbedded) AddToFlagSet(name string, flags *pflag.FlagSet) {
	f.name = name
	f.flagSet = flags

	flags.StringArrayVar(&f.target, f.Name, nil, f.Usage)
	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
