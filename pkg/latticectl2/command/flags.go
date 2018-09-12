package command

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
)

func OutputFlag(supported []printer.Format, defaultFormat printer.Format) *flags.String {
	if defaultFormat == "" {
		defaultFormat = printer.FormatTable
	}

	usage := "Set the output format of the command. Valid options: "

	var formats []string
	for _, format := range supported {
		if format == defaultFormat {
			formats = append(formats, fmt.Sprintf("%v (default)", string(format)))
		} else {
			formats = append(formats, string(format))
		}
	}

	usage += strings.Join(formats, ", ")

	return &flags.String{
		Short:   "o",
		Usage:   usage,
		Default: string(defaultFormat),
	}
}

func WatchFlag() *flags.Bool {
	return &flags.Bool{
		Required: false,
	}
}
