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
			err := DeploySystem(ctx.Client().Systems().Rollouts(systemID), types.SystemBuildID(buildID), version)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func DeploySystem(
	client client.RolloutClient,
	buildID types.SystemBuildID,
	version string,
) error {
	if buildID == "" && version == "" {
		return fmt.Errorf("must provide either build or version")
	}

	var deployID types.SystemRolloutID
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
		return err
	}

	fmt.Printf("%v\n", deployID)
	return nil
}
