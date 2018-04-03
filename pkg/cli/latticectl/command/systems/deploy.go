package systems

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/color"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems/builds"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/briandowns/spinner"
)

type DeployCommand struct {
}

func (c *DeployCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool
	var buildID string
	var version string
	cmd := &lctlcommand.SystemCommand{
		Name: "deploy",
		Flags: []command.Flag{
			output.Flag(),
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
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
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			err = DeploySystem(ctx.Client().Systems(), ctx.SystemID(), types.SystemBuildID(buildID), version, os.Stdout, format)
			if err != nil {
				//log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	return cmd.Base()
}

func DeploySystem(
	client client.SystemClient,
	systemID types.SystemID,
	buildID types.SystemBuildID,
	version string,
	writer io.Writer,
	format printer.Format,
) error {
	if buildID == "" && version == "" {
		return fmt.Errorf("must provide either build or version")
	}

	var deployID types.SystemRolloutID
	var err error
	if buildID != "" {
		if version != "" {
			log.Panic("can only provide either build or version")
			deployID, err = client.Rollouts(systemID).CreateFromBuild(buildID)
		}
	} else {
		deployID, err = client.Rollouts(systemID).CreateFromVersion(version)
	}

	if err != nil {
		return err
	}

	//TODO: Could reduce the number of requests necessary by
	// changing the behaviour of the client to return the
	// whole deploy on creation.
	deploy, err := client.Rollouts(systemID).Get(deployID)
	if err != nil {
		return err
	}

	err = builds.WatchBuild(client.SystemBuilds(systemID), deploy.BuildID, format, writer, printBuildStateDuringDeploy)
	if err != nil {
		return err
	}

	err = WatchSystem(client, systemID, format, os.Stdout, printSystemStateDuringDeploy, true)
	if err != nil {
		return err
	}

	return nil
}

func printBuildStateDuringDeploy(writer io.Writer, s *spinner.Spinner, build *types.SystemBuild) {
	switch build.State {
	case types.SystemBuildStatePending:
		s.Start()
		s.Suffix = fmt.Sprintf(" Build pending for version: %s...", color.ID(string(build.Version)))
	case types.SystemBuildStateRunning:
		s.Start()
		s.Suffix = fmt.Sprintf(" Building version: %s...", color.ID(string(build.Version)))
	case types.SystemBuildStateSucceeded:
		s.Stop()

		fmt.Fprint(writer, color.BoldHiSuccess("✓ %s built successfully! Now deploying...\n", string(build.Version)))
	case types.SystemBuildStateFailed:
		s.Stop()

		var componentErrors [][]string

		for serviceName, service := range build.Services {
			for componentName, component := range service.Components {
				if component.State == types.ComponentBuildStateFailed {
					componentErrors = append(componentErrors, []string{
						fmt.Sprintf("%s:%s", serviceName, componentName),
						string(*component.FailureMessage),
					})
				}
			}
		}

		builds.PrintBuildFailure(writer, string(build.Version), componentErrors)
	}
}

func printSystemStateDuringDeploy(writer io.Writer, s *spinner.Spinner, system *types.System) {
	switch system.State {
	case types.SystemStateScaling:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is scaling...", color.ID(string(system.ID)))
	case types.SystemStateUpdating:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is updating...", color.ID(string(system.ID)))
	case types.SystemStateDeleting:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is terminating...", color.ID(string(system.ID)))
	case types.SystemStateStable:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiSuccess("✓ Rollout for system %s has succeeded.\n", string(system.ID)))
	case types.SystemStateFailed:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiFailure("✘ Rollout for system %s has failed.\n", string(system.ID)))

		var serviceErrors [][]string

		for serviceName, service := range system.Services {
			if service.State == types.ServiceStateFailed {
				serviceErrors = append(serviceErrors, []string{
					fmt.Sprintf("%s", serviceName),
					string(*service.FailureMessage),
				})
			}
		}

		printSystemFailure(writer, system.ID, serviceErrors)
	}
}
