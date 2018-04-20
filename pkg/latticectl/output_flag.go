package latticectl

import (
	"fmt"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

type OutputFlag struct {
	Name             string
	Short            *string
	SupportedFormats []printer.Format
	value            string
}

func (f *OutputFlag) Flag() cli.Flag {
	name := f.Name
	if name == "" {
		name = "output"
	}

	short := "o"
	if f.Short != nil {
		short = *f.Short
	}

	return &cli.StringFlag{
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
