package cli

import (
	"fmt"

	"github.com/spf13/pflag"
)

// EmbeddedFlag is a Flag that allows you to parse multiple values from the same flag.
// For example, if c is an embedded flag with two string flags, bar and buzz,
// you could say --c "bar=hello,buzz=world".
type EmbeddedFlag struct {
	Name      string
	Required  bool
	Short     string
	Usage     string
	Flags     Flags
	Delimiter string
	target    []string
}

func (f *EmbeddedFlag) GetName() string {
	return f.Name
}

func (f *EmbeddedFlag) IsRequired() bool {
	return f.Required
}

func (f *EmbeddedFlag) GetShort() string {
	return f.Short
}

func (f *EmbeddedFlag) GetUsage() string {
	return f.Usage
}

func (f *EmbeddedFlag) Validate() error {
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

func (f *EmbeddedFlag) GetTarget() interface{} {
	return nil
}

func (f *EmbeddedFlag) Parse() func() error {
	return f.parse
}

func (f *EmbeddedFlag) parse() error {
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

func (f *EmbeddedFlag) AddToFlagSet(flags *pflag.FlagSet) {
	flags.StringArrayVar(&f.target, f.Name, nil, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}

type DelayedEmbeddedFlag struct {
	Name        string
	Required    bool
	Short       string
	Usage       string
	Flags       map[string]Flags
	Delimiter   string
	FlagChooser func() (string, error)
	target      []string
}

func (f *DelayedEmbeddedFlag) GetName() string {
	return f.Name
}

func (f *DelayedEmbeddedFlag) IsRequired() bool {
	return f.Required
}

func (f *DelayedEmbeddedFlag) GetShort() string {
	return f.Short
}

func (f *DelayedEmbeddedFlag) GetUsage() string {
	return f.Usage
}

func (f *DelayedEmbeddedFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.FlagChooser == nil {
		return fmt.Errorf("FlagChooser cannot be nil")
	}

	return nil
}

func (f *DelayedEmbeddedFlag) GetTarget() interface{} {
	return nil
}

func (f *DelayedEmbeddedFlag) Parse() func() error {
	return f.parse
}

func (f *DelayedEmbeddedFlag) parse() error {
	choice, err := f.FlagChooser()
	if err != nil {
		return err
	}

	flags, ok := f.Flags[choice]
	if !ok {
		return fmt.Errorf("invalid flag choice %v", choice)
	}

	embedded := &EmbeddedFlag{
		Name:      f.Name,
		Required:  f.Required,
		Short:     f.Short,
		Usage:     f.Usage,
		Flags:     flags,
		Delimiter: f.Delimiter,
		target:    f.target,
	}

	if err := embedded.Validate(); err != nil {
		return err
	}

	return embedded.parse()
}

func (f *DelayedEmbeddedFlag) AddToFlagSet(flags *pflag.FlagSet) {
	flags.StringArrayVar(&f.target, f.Name, nil, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
