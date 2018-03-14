package command

import (
	"fmt"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
)

type OutputFlag struct {
	Name             string
	Short            *string
	SupportedFormats []printer.Format
	value            string
}

func (f *OutputFlag) Flag() command.Flag {
	name := f.Name
	if name == "" {
		name = "output"
	}

	short := "o"
	if f.Short != nil {
		short = *f.Short
	}

	return &command.StringFlag{
		Name:   name,
		Short:  short,
		Target: &f.value,
	}
}

func (f *OutputFlag) Value() (printer.Format, error) {
	if f.value == "" {
		return printer.FormatDefault, nil
	}

	value := printer.Format(f.value)
	for _, format := range f.SupportedFormats {
		if value == format {
			return value, nil
		}
	}

	return "", fmt.Errorf("unsupported format: %v", value)
}
