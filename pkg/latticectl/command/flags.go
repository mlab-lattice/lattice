package command

import (
	"strings"

	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

const (
	ConfigFlagName  = "config"
	ContextFlagName = "context"
	OutputFlagName  = "output"
	SidecarFlagName = "sidecar"
	SystemFlagName  = "system"
	WatchFlagName   = "watch"
)

// ConfigFlag returns the canonical flag for specifying a config file.
func ConfigFlag(target *string) *flags.String {
	return &flags.String{Target: target}
}

// ContextFlag returns the canonical flag for specifying a context.
func ContextFlag(target *string) *flags.String {
	return &flags.String{Target: target}
}

// OutputFlag returns the canonical flag for specifying an output format.
func OutputFlag(target *string, supported []printer.Format, defaultFormat printer.Format) *flags.String {
	usage := "Set the output format of the command. Valid options: "

	var formats []string
	for _, format := range supported {
		formats = append(formats, string(format))
	}

	usage += strings.Join(formats, ", ")

	return &flags.String{
		Short:   "o",
		Usage:   usage,
		Default: string(defaultFormat),
		Target:  target,
	}
}

// SidecarFlag returns the canonical flag for specifying a sidecar.
func SidecarFlag(target *string) *flags.String {
	return &flags.String{Target: target}
}

// SystemFlag returns the canonical flag for specifying a system.
func SystemFlag(target *string) *flags.String {
	return &flags.String{Target: target}
}

// WatchFlag returns the canonical flag for specifying the command is a watch	.
func WatchFlag(target *bool) *flags.Bool {
	return &flags.Bool{
		Short:    "w",
		Required: false,
		Target:   target,
	}
}
