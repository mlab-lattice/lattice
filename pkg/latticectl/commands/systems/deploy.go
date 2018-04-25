package systems

import (
	"fmt"
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/systems/builds"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"github.com/briandowns/spinner"
)

type DeployCommand struct {
}

func (c *DeployCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target: &watch,
	}
	var buildID string
	var version string
	cmd := &latticectl.SystemCommand{
		Name: "deploy",
		Flags: []cli.Flag{
			output.Flag(),
			watchFlag.Flag(),
			&cli.StringFlag{
				Name:     "build",
				Required: false,
				Target:   &buildID,
			},
			&cli.StringFlag{
				Name:     "version",
				Required: false,
				Target:   &version,
			},
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			err = DeploySystem(ctx.Client().Systems(), ctx.SystemID(), v1.BuildID(buildID), v1.SystemVersion(version), os.Stdout, format, watch)
			if err != nil {
				//log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	return cmd.Base()
}

func DeploySystem(
	client v1client.SystemClient,
	systemID v1.SystemID,
	buildID v1.BuildID,
	version v1.SystemVersion,
	writer io.Writer,
	format printer.Format,
	watch bool,
) error {
	if buildID == "" && version == "" {
		return fmt.Errorf("must provide either build or version")
	}

	var deploy *v1.Deploy
	var err error
	var definition string

	if buildID != "" {
		if version != "" {
			log.Panic("can only provide either build or version")
		}
		definition = fmt.Sprintf("build %s", color.ID(string(buildID)))
		deploy, err = client.Deploys(systemID).CreateFromBuild(buildID)
	} else {
		definition = fmt.Sprintf("version %s", color.ID(string(version)))
		deploy, err = client.Deploys(systemID).CreateFromVersion(version)
	}

	if err != nil {
		return err
	}

	if watch {
		err = builds.WatchBuild(client.Builds(systemID), deploy.BuildID, format, writer, printBuildStateDuringDeploy)
		if err != nil {
			return err
		}

		err = WatchSystem(client, systemID, format, os.Stdout, printSystemStateDuringDeploy, true)
		if err != nil {
			return err
		}
	} else {
		fmt.Fprintf(writer, "\nDeploying %s for system %s. Deploy ID: %s\n\n", definition, color.ID(string(systemID)), color.ID(string(deploy.ID)))
		fmt.Fprint(writer, "To watch deploy, run:\n\n")
		fmt.Fprintf(writer, "    lattice system:deploys:status -w --deploy %s\n", string(deploy.ID))
	}

	return nil
}

func printBuildStateDuringDeploy(writer io.Writer, s *spinner.Spinner, build *v1.Build) {
	switch build.State {
	case v1.BuildStatePending:
		s.Start()
		s.Suffix = fmt.Sprintf(" Build pending for version: %s...", color.ID(string(build.Version)))
	case v1.BuildStateRunning:
		s.Start()
		s.Suffix = fmt.Sprintf(" Building version: %s...", color.ID(string(build.Version)))
	case v1.BuildStateSucceeded:
		s.Stop()

		fmt.Fprint(writer, color.BoldHiSuccess("✓ %s built successfully! Now deploying...\n", string(build.Version)))
	case v1.BuildStateFailed:
		s.Stop()

		var componentErrors [][]string

		for serviceName, service := range build.Services {
			for componentName, component := range service.Components {
				if component.State == v1.ComponentBuildStateFailed {
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

func printSystemStateDuringDeploy(writer io.Writer, s *spinner.Spinner, system *v1.System, services []v1.Service) {
	switch system.State {
	case v1.SystemStateScaling:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is scaling...", color.ID(string(system.ID)))
	case v1.SystemStateUpdating:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is updating...", color.ID(string(system.ID)))
	case v1.SystemStateDeleting:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is terminating...", color.ID(string(system.ID)))
	case v1.SystemStateStable:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiSuccess("✓ Rollout for system %s has succeeded.\n", string(system.ID)))
	case v1.SystemStateFailed:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiFailure("✘ Rollout for system %s has failed.\n", string(system.ID)))

		var serviceErrors [][]string

		for _, service := range services {
			if service.State == v1.ServiceStateFailed {
				message := "unknown"
				if service.FailureInfo != nil {
					message = service.FailureInfo.Message
				}

				serviceErrors = append(serviceErrors, []string{
					service.Path.String(),
					message,
				})
			}
		}

		printSystemFailure(writer, system.ID, serviceErrors)
	}
}
