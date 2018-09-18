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

	values []string
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
	for name, flag := range f.Flags {
		if flag.GetShort() != "" {
			return fmt.Errorf("embedded flag %v cannot have a short", name)
		}
	}

	return nil
}

func (f *Embedded) Value() interface{} {
	return f.Flags
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
	for _, value := range f.values {
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

func (f *Embedded) AddToFlagSet(name string, flags *pflag.FlagSet) {
	flags.StringArrayVar(&f.values, name, nil, f.Usage)

	if f.Required {
		markFlagRequired(name, flags)
	}
}

//
//type EmbeddedStringChoice struct {
//	ChoiceFlag *String
//	Flags      map[string]cli.Flags
//
//	result cli.Flags
//}
//
//func (f *EmbeddedStringChoice) IsRequired() bool {
//	return f.ChoiceFlag.IsRequired()
//}
//
//func (f *EmbeddedStringChoice) GetShort() string {
//	return f.ChoiceFlag.GetShort()
//}
//
//func (f *EmbeddedStringChoice) GetUsage() string {
//	return f.ChoiceFlag.GetUsage()
//}
//
//func (f *EmbeddedStringChoice) Value() interface{} {
//	return f.result
//}
//
//func (f *EmbeddedStringChoice) Parse() func() error {
//	return f.parse
//}
//
//func (f *EmbeddedStringChoice) parse() error {
//	choice := f.ChoiceFlag.Value().(string)
//	flags, ok := f.Flags[choice]
//	if !ok {
//		return fmt.Errorf("invalid flag choice: %v", choice)
//	}
//
//	fs := &pflag.FlagSet{}
//	var dashedValues []string
//	for name, flag := range flags {
//		flag.AddToFlagSet(name, fs)
//		dashedValues = append(dashedValues, fmt.Sprintf("--%v", name))
//	}
//
//	err := fs.Parse(dashedValues)
//	if err != nil {
//		return err
//	}
//
//	f.result = flags
//	return nil
//}
//
//func (f *EmbeddedStringChoice) AddToFlagSet(name string, flags *pflag.FlagSet) {
//	f.ChoiceFlag.AddToFlagSet(name, flags)
//}
