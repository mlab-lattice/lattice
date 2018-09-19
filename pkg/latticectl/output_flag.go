package latticectl

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

type OutputFlag struct {
	Name             string
	Short            string
	SupportedFormats []printer.Format
	DefaultFormat    printer.Format
	Usage            string
	value            string
}

func (f *OutputFlag) Flag() cli.Flag {
	name := f.Name
	if name == "" {
		name = "output"
	}

	short := "o"
	if f.Short != "" {
		short = f.Short
	}

	// You can set the default format per command, but the overall default is table
	if f.DefaultFormat == "" {
		f.DefaultFormat = printer.FormatTable
	}

	usage := "Set the output format of the command. Valid options: "

	var formats []string
	for _, format := range f.SupportedFormats {
		if format == f.DefaultFormat {
			formats = append(formats, string(format)+" (default)")
		} else {
			formats = append(formats, string(format))
		}
	}

	usage += strings.Join(formats, ", ")

	if f.Usage != "" {
		usage = f.Usage
	}

	return &flags.String{
		Name:   name,
		Short:  short,
		Usage:  usage,
		Target: &f.value,
	}
}

func (f *OutputFlag) Value() (printer.Format, error) {
	if f.value == "" {
		return f.DefaultFormat, nil
	}

	value := printer.Format(f.value)
	for _, format := range f.SupportedFormats {
		if value == format {
			return value, nil
		}
	}

	return "", fmt.Errorf("unsupported format: %v", value)
}
