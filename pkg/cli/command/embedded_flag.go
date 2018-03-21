package command

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

type EmbeddedFlag struct {
	Name      string
	Required  bool
	Short     string
	Usage     string
	Flags     Flags
	Delimiter string
	target    string
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
	return f.parseEmbeddedFlag
}

func (f *EmbeddedFlag) parseEmbeddedFlag() error {
	flags := &pflag.FlagSet{}
	for _, flag := range f.Flags {
		flag.AddToFlagSet(flags)
	}

	// default the delimiter to ,
	delimiter := ","
	if f.Delimiter != "" {
		delimiter = f.Delimiter
	}

	values := strings.Split(f.target, delimiter)

	var dashedValues []string
	for _, value := range values {
		dashedValues = append(dashedValues, fmt.Sprintf("--%v", value))
	}

	err := flags.Parse(dashedValues)
	if err != nil {
		return nil
	}

	for _, flag := range f.Flags {
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
	flags.StringVarP(&f.target, f.Name, f.Short, "", f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
