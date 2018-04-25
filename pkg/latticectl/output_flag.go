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
	Usage            string
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

	usage := "Set the output format of the command. Options are table (which is the default) and json."
	if f.Usage != "" {
		usage = f.Usage
	}

	return &cli.StringFlag{
		Name:   name,
		Short:  short,
		Usage:  usage,
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
