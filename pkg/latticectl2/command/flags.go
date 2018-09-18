package command

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
)

const (
	ConfigFlagName  = "config"
	ContextFlagName = "Context"
	OutputFlagName  = "output"
	SystemFlagName  = "system"
	WatchFlagName   = "watch"
)

func ConfigFlag(target *string) *flags.String {
	return &flags.String{Target: target}
}

func ContextFlag(target *string) *flags.String {
	return &flags.String{Target: target}
}

func OutputFlag(target *string, supported []printer.Format, defaultFormat printer.Format) *flags.String {
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
		Target:  target,
	}
}

func SystemFlag(target *string) *flags.String {
	return &flags.String{
		Required: false,
		Target:   target,
	}
}

func WatchFlag(target *bool) *flags.Bool {
	return &flags.Bool{
		Short:    "w",
		Required: false,
		Target:   target,
	}
}
