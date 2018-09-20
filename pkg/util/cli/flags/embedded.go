package flags

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/util/cli"

	"github.com/spf13/pflag"
)

// Embedded is a Flag that allows you to parse multiple values from the same flag.
// For example, if c is an embedded flag with two string flags, bar and buzz,
// you could say --c "bar=hello,buzz=world".
type Embedded struct {
	Name     string
	Required bool
	Short    string
	Usage    string
	Flags    cli.Flags
	target   []string
}

func (f *Embedded) GetName() string {
	return f.Name
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

func (f *Embedded) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	for _, flag := range f.Flags {
		if flag.GetShort() != "" {
			return fmt.Errorf("embedded flag %v cannot have a Short", flag.GetName())
		}

		if err := flag.Validate(); err != nil {
			return fmt.Errorf("error validating embedded flag %v: %v", flag.GetName(), err)
		}
	}

	return nil
}

func (f *Embedded) GetTarget() interface{} {
	return nil
}

func (f *Embedded) Parse() func() error {
	return f.parse
}

func (f *Embedded) parse() error {
	flags := &pflag.FlagSet{}
	for _, flag := range f.Flags {
		flag.AddToFlagSet(flags)
	}

	var dashedValues []string
	for _, value := range f.target {
		dashedValues = append(dashedValues, fmt.Sprintf("--%v", value))
	}

	err := flags.Parse(dashedValues)
	if err != nil {
		return err
	}

	for _, flag := range f.Flags {
		if flag.IsRequired() && !flags.Changed(flag.GetName()) {
			return fmt.Errorf("missing requrired flag: %v", flag.GetName())
		}

		parser := flag.Parse()
		if parser != nil {
			err := parser()
			if err != nil {
				return fmt.Errorf("error parsing embedded flag %v: %v", flag.GetName(), err)
			}
		}
	}

	return nil
}

func (f *Embedded) AddToFlagSet(flags *pflag.FlagSet) {
	flags.StringArrayVar(&f.target, f.Name, nil, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
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
	target      []string
}

func (f *DelayedEmbedded) GetName() string {
	return f.Name
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

func (f *DelayedEmbedded) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.FlagChooser == nil {
		return fmt.Errorf("FlagChooser cannot be nil")
	}

	return nil
}

func (f *DelayedEmbedded) GetTarget() interface{} {
	return nil
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
		Name:     f.Name,
		Required: f.Required,
		Short:    f.Short,
		Usage:    f.Usage,
		Flags:    flags,
		target:   f.target,
	}

	if err := embedded.Validate(); err != nil {
		return err
	}

	return embedded.parse()
}

func (f *DelayedEmbedded) AddToFlagSet(flags *pflag.FlagSet) {
	flags.StringArrayVar(&f.target, f.Name, nil, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
