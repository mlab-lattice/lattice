package deploys

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"k8s.io/apimachinery/pkg/util/wait"
)

// GetDeploysSupportedFormats is the list of printer.Formats supported
// by the GetDeploy function.
var GetDeploysSupportedFormats = []printer.Format{
	printer.FormatJSON,
	printer.FormatTable,
}

type StatusCommand struct {
}

func (c *StatusCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: GetDeploysSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target: &watch,
	}

	cmd := &latticectl.DeployCommand{
		Name: "status",
		Flags: cli.Flags{
			output.Flag(),
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.DeployCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().Deploys(ctx.SystemID())

			if watch {
				WatchDeploy(c, ctx.DeployID(), format, os.Stdout)
			}

			err = GetDeploy(c, ctx.DeployID(), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetDeploy(client v1client.SystemDeployClient, deployID v1.DeployID, format printer.Format, writer io.Writer) error {
	deploy, err := client.Get(deployID)
	if err != nil {
		return err
	}

	p := deploysPrinter([]v1.Deploy{*deploy}, format)
	p.Print(writer)
	return nil
}

func WatchDeploy(client v1client.SystemDeployClient, deployID v1.DeployID, format printer.Format, writer io.Writer) {
	deploys := make(chan *v1.Deploy)

	lastHeight := 0
	var b bytes.Buffer

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			deploy, err := client.Get(deployID)
			if err != nil {
				return false, err
			}

			deploys <- deploy
			return false, nil
		},
	)

	for deploy := range deploys {
		p := deploysPrinter([]v1.Deploy{*deploy}, format)
		lastHeight = p.Overwrite(b, lastHeight)
	}
}
