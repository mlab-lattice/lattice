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

type ListVersionsCommand struct {
}

func (c *ListVersionsCommand) Base() (*latticectl.BaseCommand, error) {
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
			// FIXME :: get versions instead.
			err := ListVersions(ctx.Client().Systems())
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func ListVersions(client client.SystemClient) error {
	// FIXME :: simple change once new api is available.
	versions, err := client.List()
	if err != nil {
		log.Panic(err)
	}

	for _, version := range versions {
		// FIXME :: will just be the string.
		fmt.Printf("%v\n", version.DefinitionURL)
	}
	return nil
}
