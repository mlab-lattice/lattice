package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems/teardowns"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

type TeardownCommand struct {
}

func (c *TeardownCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: teardowns.ListTeardownsSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.SystemCommand{
		Name: "teardown",
		Flags: []command.Flag{
			output.Flag(),
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			systemID := ctx.SystemID()
			TeardownSystem(ctx.Client().Systems().Teardowns(systemID))
		},
	}

	return cmd.Base()
}

func TeardownSystem(
	client client.TeardownClient,
) {
	// TODO :: Add watch of this. Same with deploy / build - link to behavior of teardowns/get.go etc
	teardown, err := client.Create()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", teardown)
}
