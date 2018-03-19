package builds

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListBuildsSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.BuildCommand{
		Name: "status",
		Flags: command.Flags{
			output.Flag(),
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx lctlcommand.BuildCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().SystemBuilds(ctx.SystemID())

			if watch {
				WatchBuild(c, ctx.BuildID(), format, os.Stdout)
			}

			err = GetBuild(c, ctx.BuildID(), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetBuild(client client.SystemBuildClient, buildID types.SystemBuildID, format printer.Format, writer io.Writer) error {
	build, err := client.Get(buildID)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", build)
	return nil
}

func WatchBuild(client client.SystemBuildClient, buildID types.SystemBuildID, format printer.Format, writer io.Writer) {
	build, err := client.Get(buildID)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", build)
}
