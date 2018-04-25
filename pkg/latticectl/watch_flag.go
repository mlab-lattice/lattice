package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type WatchFlag struct {
	Name   string
	Short  *string
	Usage  string
	Target *bool
}

func (f *WatchFlag) Flag() cli.Flag {
	name := f.Name
	if name == "" {
		name = "watch"
	}

	short := "w"
	if f.Short != nil {
		short = *f.Short
	}

	usage := "If the watch flag is set, the output will update every 5 seconds."
	if f.Usage != "" {
		usage = f.Usage
	}

	return &cli.BoolFlag{
		Name:    name,
		Short:   short,
		Usage:   usage,
		Default: false,
		Target:  f.Target,
	}
}
