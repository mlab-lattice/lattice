package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type DeployCommand struct {
}

func (c *DeployCommand) Base() (*latticectl.BaseCommand, error) {
	var buildID string
	var version string
	cmd := &lctlcommand.SystemCommand{
		Name: "deploy",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:     "build",
				Required: false,
				Target:   &buildID,
			},
			&command.StringFlag{
				Name:     "version",
				Required: false,
				Target:   &version,
			},
		},
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			systemID := ctx.SystemID()
			DeploySystem(ctx.Client().Systems().Rollouts(systemID), types.BuildID(buildID), version)
		},
	}

	return cmd.Base()
}

func DeploySystem(
	client client.RolloutClient,
	buildID types.BuildID,
	version string,
) {
	if buildID == "" && version == "" {
		log.Panic("must provide either build or version")
	}

	var deployID types.DeployID
	var err error
	if buildID != "" {
		if version != "" {
			log.Panic("can only provide either build or version")
			deployID, err = client.CreateFromBuild(buildID)
		}
	} else {
		deployID, err = client.CreateFromVersion(version)
	}

	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deployID)
}
