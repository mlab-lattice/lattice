package systems

import (
	"fmt"
	"log"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
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
	output := &latticectl.OutputFlag{
		SupportedFormats: ListVersionsSupportedFormats,
	}

	cmd := &latticectl.SystemCommand{
		Name: "versions",
		Flags: []cli.Flag{
			output.Flag(),
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			// FIXME :: get versions instead.
			err := ListVersions(ctx.Client().Systems())
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func ListVersions(client v1client.SystemClient) error {
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
