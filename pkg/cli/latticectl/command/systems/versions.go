package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems/teardowns"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

// ListVersionsSupportedFormats is the list of printer.Formats supported
// by the ListTeardowns function.
var ListVersionsSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

type ListVersionsCommand struct {
}

func (c *ListVersionsCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: teardowns.ListTeardownsSupportedFormats,
	}

	cmd := &lctlcommand.SystemCommand{
		Name: "versions",
		Flags: []command.Flag{
			output.Flag(),
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
