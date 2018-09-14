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
	// get the deploy's build
	var build *latticev1.Build
	if deploy.Spec.Build != nil {
		var err error
		build, err = c.buildLister.Builds(deploy.Namespace).Get(string(*deploy.Status.BuildID))
		if err != nil {
			return err
		}
	} else {
		var err error
		deploy, build, err = c.syncAcceptedBuildlessDeploy(deploy)
		if err != nil {
			return err
		}
	}

	build, err := c.addBuildOwnerReference(deploy, build)
	if err != nil {
		return err
	}

	switch build.Status.State {
	// if the build has not reached a terminal state, there's nothing to do yet
	case latticev1.BuildStatePending, latticev1.BuildStateAccepted, latticev1.BuildStateRunning:
		return nil

	// if the build failed, fail the deploy as well
	case latticev1.BuildStateFailed:
		_, err := c.updateDeployStatus(
			deploy,
			latticev1.DeployStateFailed,
			fmt.Sprintf("%v failed", build.Description(c.namespacePrefix)),
			nil,
			deploy.Status.BuildID,
		)
		if err != nil {
			return err
		}

		// release the deploy's lock so other deploys can deploy along this path
		err = c.releaseDeployLock(deploy)
		return err

	case latticev1.BuildStateSucceeded:
		return c.syncAcceptedDeployWithSuccessfulBuild(deploy, build)

	default:
		return fmt.Errorf("%v in unexpected state %v", build.Description(c.namespacePrefix), build.Status.State)
	}
}

func (c *Controller) syncAcceptedBuildlessDeploy(deploy *latticev1.Deploy) (*latticev1.Deploy, *latticev1.Build, error) {
	// If we've already created a build and updated the status of the deploy with it, use that build ID
	if deploy.Status.BuildID != nil {
		build, err := c.buildLister.Builds(deploy.Namespace).Get(string(*deploy.Status.BuildID))
		if err != nil {
			return nil, nil, err
		}

		return deploy, build, nil
	}

	// Otherwise create the build and update the deploy's status
	build := &latticev1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: deploy.Namespace,
			Name:      uuid.NewV4().String(),
		},
		Spec: latticev1.BuildSpec{
			Version: deploy.Spec.Version,
			Path:    deploy.Spec.Path,
		},
	}

	build, err := c.latticeClient.LatticeV1().Builds(deploy.Namespace).Create(build)
	if err != nil {
		return nil, nil, err
	}

	buildID := v1.BuildID(build.Name)
	deploy, err = c.updateDeployStatus(deploy, latticev1.DeployStateAccepted, "", nil, &buildID)
	if err != nil {
		return nil, nil, err
	}

	return deploy, build, nil
}

func (c *Controller) syncAcceptedDeployWithSuccessfulBuild(deploy *latticev1.Deploy, build *latticev1.Build) error {
	system, err := c.getSystem(deploy.Namespace)
	if err != nil {
		return err
	}

	version := v1.Version("unknown")
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

	// loop through all of the workloads and seed there artifacts into the artifacts
	// tree
	err = nil
	artifacts := latticev1.NewSystemSpecWorkloadBuildArtifacts()
	seedArtifacts := func(p tree.Path, _ definitionv1.Workload, info *resolver.ResolutionInfo) tree.WalkContinuation {
		// first get the artifacts for the main container
		workloadInfo, ok := build.Status.Workloads[p]
		if !ok {
			err = fmt.Errorf(
				"%v had workload %v but no information about it",
				build.Description(c.namespacePrefix),
				p.String(),
			)
			return tree.HaltWalk
		}

		mainContainerBuild, ok := build.Status.ContainerBuildStatuses[workloadInfo.MainContainer]
		if !ok {
			err = fmt.Errorf(
				"%v had workload %v container build %v but no information about it",
				build.Description(c.namespacePrefix),
				p.String(),
				workloadInfo.MainContainer,
			)
			return tree.HaltWalk
		}

		if mainContainerBuild.Artifacts == nil {
			err = fmt.Errorf(
				"%v had workload %v container build %v but artifacts are nil",
				build.Description(c.namespacePrefix),
				p.String(),
				workloadInfo.MainContainer,
			)
			return tree.HaltWalk
		}

		workloadArtifacts := latticev1.WorkloadContainerBuildArtifacts{
			MainContainer: *mainContainerBuild.Artifacts,
			Sidecars:      make(map[string]latticev1.ContainerBuildArtifacts),
		}

		// get the artifacts for all of the sidecars
		for sidecar, sidecarBuild := range workloadInfo.Sidecars {
			containerBuild, ok := build.Status.ContainerBuildStatuses[sidecarBuild]
			if !ok {
				err = fmt.Errorf(
					"%v had workload %v container build %v but no information about it",
					build.Description(c.namespacePrefix),
					p.String(),
					sidecarBuild,
				)
				return tree.HaltWalk
			}

			if containerBuild.Artifacts == nil {
				err = fmt.Errorf(
					"%v had workload %v container build %v but artifacts are nil",
					build.Description(c.namespacePrefix),
					p.String(),
					sidecarBuild,
				)
				return tree.HaltWalk
			}

			workloadArtifacts.Sidecars[sidecar] = *containerBuild.Artifacts
		}

		artifacts.Insert(p, workloadArtifacts)
		return tree.ContinueWalk
	}

	build.Status.Definition.V1().Workloads(seedArtifacts)
	if err != nil {
		return err
	}

	spec := system.Spec.DeepCopy()
	if spec.Definition == nil {
		spec.Definition = resolver.NewComponentTree()
	}

	if spec.WorkloadBuildArtifacts == nil {
		spec.WorkloadBuildArtifacts = latticev1.NewSystemSpecWorkloadBuildArtifacts()
	}

	path := tree.RootPath()
	if build.Spec.Path != nil {
		path = *build.Spec.Path
	}

	spec.Definition.ReplacePrefix(path, build.Status.Definition)
	spec.WorkloadBuildArtifacts.ReplacePrefix(path, artifacts)

	_, err = c.updateSystemSpec(system, spec)
	if err != nil {
		return err
	}

	_, err = c.updateDeployStatus(deploy, latticev1.DeployStateInProgress, "", nil, deploy.Status.BuildID)
	return err
}
