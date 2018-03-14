package deploys

import (
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListDeploysSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.DeployCommand{
		Name: "get",
		Flags: command.Flags{
			output.Flag(),
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx lctlcommand.DeployCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().Rollouts(ctx.SystemID())

			getFunc := func(client client.RolloutClient) ([]types.SystemRollout, error) {
				rollout, err := client.Get(ctx.DeployID())
				if err != nil {
					return nil, err
				}
				return []types.SystemRollout{*rollout}, nil
			}

			if watch {
				WatchDeploys(getFunc, c, format, os.Stdout)
				return
			}

			ListDeploys(getFunc, c, format, os.Stdout)
		},
	}

	return cmd.Base()
}