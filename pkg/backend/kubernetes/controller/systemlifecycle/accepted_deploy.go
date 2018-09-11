package systemlifecycle

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (c *Controller) syncAcceptedDeploy(deploy *latticev1.Deploy) error {
	if deploy.Spec.Version == nil && deploy.Spec.Build == nil {
		return fmt.Errorf("%v had neither version nor build id", deploy.Description(c.namespacePrefix))
	}

	var build *latticev1.Build
	var err error
	switch {
	case deploy.Spec.Version != nil:
		// If we've already created a build and updated the status of the deploy with it, use that build ID
		if deploy.Status.BuildID != nil {
			build, err = c.buildLister.Builds(deploy.Namespace).Get(string(*deploy.Status.BuildID))
			if err != nil {
				return err
			}
			break
		}

		// Otherwise create the build and update the deploy's status
		build := &latticev1.Build{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: deploy.Namespace,
				Name:      uuid.NewV4().String(),
			},
			Spec: latticev1.BuildSpec{
				Version: deploy.Spec.Version.Version,
				Path:    deploy.Spec.Version.Path,
			},
		}

		build, err = c.latticeClient.LatticeV1().Builds(deploy.Namespace).Create(build)
		if err != nil {
			return err
		}

		buildID := v1.BuildID(build.Name)
		deploy, err = c.updateDeployStatus(deploy, latticev1.DeployStateAccepted, "", &buildID)
		if err != nil {
			return err
		}

	case deploy.Spec.Build != nil:
		build, err = c.buildLister.Builds(deploy.Namespace).Get(string(*deploy.Status.BuildID))
		if err != nil {
			return err
		}
	}

	switch build.Status.State {
	case latticev1.BuildStatePending, latticev1.BuildStateAccepted, latticev1.BuildStateRunning:
		return nil

	case latticev1.BuildStateFailed:
		_, err := c.updateDeployStatus(
			deploy,
			latticev1.DeployStateFailed,
			fmt.Sprintf("%v failed", build.Description(c.namespacePrefix)),
			deploy.Status.BuildID,
		)
		if err != nil {
			return err
		}

		return c.relinquishDeployOwningActionClaim(deploy)

	case latticev1.BuildStateSucceeded:
		system, err := c.getSystem(deploy.Namespace)
		if err != nil {
			return err
		}

		version := v1.SystemVersion("unknown")
		if label, ok := deploy.DefinitionVersionLabel(); ok {
			version = label
		}

		buildID := v1.BuildID("unknown")
		if label, ok := deploy.BuildIDLabel(); ok {
			buildID = label
		}

		deployID := v1.DeployID(deploy.Name)

		system, err = c.updateSystemLabels(system, &version, &deployID, &buildID)
		if err != nil {
			return err
		}

		artifacts := latticev1.NewSystemSpecWorkloadBuildArtifacts()
		err = nil
		build.Status.Definition.V1().Workloads(func(path tree.Path, workload definitionv1.Workload, info *resolver.ResolutionInfo) bool {
			workloadInfo, ok := build.Status.Workloads[path]
			if !ok {
				err = fmt.Errorf(
					"%v had workload %v but no information about it",
					build.Description(c.namespacePrefix),
					path.String(),
				)
				return false
			}

			mainContainerBuild, ok := build.Status.ContainerBuildStatuses[workloadInfo.MainContainer]
			if !ok {
				err = fmt.Errorf(
					"%v had workload %v container build %v but no information about it",
					build.Description(c.namespacePrefix),
					path.String(),
					workloadInfo.MainContainer,
				)
				return false
			}

			if mainContainerBuild.Artifacts == nil {
				err = fmt.Errorf(
					"%v had workload %v container build %v but artifacts are nil",
					build.Description(c.namespacePrefix),
					path.String(),
					workloadInfo.MainContainer,
				)
				return false
			}

			workloadArtifacts := latticev1.WorkloadContainerBuildArtifacts{
				MainContainer: *mainContainerBuild.Artifacts,
				Sidecars:      make(map[string]latticev1.ContainerBuildArtifacts),
			}

			for sidecar, sidecarBuild := range workloadInfo.Sidecars {
				containerBuild, ok := build.Status.ContainerBuildStatuses[sidecarBuild]
				if !ok {
					err = fmt.Errorf(
						"%v had workload %v container build %v but no information about it",
						build.Description(c.namespacePrefix),
						path.String(),
						sidecarBuild,
					)
					return false
				}

				if containerBuild.Artifacts == nil {
					err = fmt.Errorf(
						"%v had workload %v container build %v but artifacts are nil",
						build.Description(c.namespacePrefix),
						path.String(),
						sidecarBuild,
					)
					return false
				}

				workloadArtifacts.Sidecars[sidecar] = *containerBuild.Artifacts
			}

			artifacts.Insert(path, workloadArtifacts)
			return true
		})
		if err != nil {
			return err
		}

		spec := system.Spec.DeepCopy()
		spec.Definition.ReplacePrefix(build.Spec.Path, build.Status.Definition)
		spec.WorkloadBuildArtifacts.ReplacePrefix(build.Spec.Path, artifacts)

		_, err = c.updateSystemSpec(system, spec)
		if err != nil {
			return err
		}

		_, err = c.updateDeployStatus(deploy, latticev1.DeployStateInProgress, "", deploy.Status.BuildID)
		return err

	default:
		return fmt.Errorf("%v in unexpected state %v", build.Description(c.namespacePrefix), build.Status.State)
	}
}
