package v1

import (
	"io"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type Interface interface {
	Systems() SystemBackend
}

type SystemBackend interface {
	Define(id v1.SystemID, url string) (*v1.System, error)
	List() ([]v1.System, error)
	Get(v1.SystemID) (*v1.System, error)
	Delete(v1.SystemID) error

	Builds(v1.SystemID) SystemBuildBackend
	Deploys(v1.SystemID) SystemDeployBackend
	Jobs(v1.SystemID) SystemJobBackend
	NodePools(v1.SystemID) SystemNodePoolBackend
	Secrets(v1.SystemID) SystemSecretBackend
	Services(v1.SystemID) SystemServiceBackend
	Teardowns(v1.SystemID) SystemTeardownBackend
}

type SystemBuildBackend interface {
	CreateFromPath(tree.Path) (*v1.Build, error)
	CreateFromVersion(v1.Version) (*v1.Build, error)
	List() ([]v1.Build, error)
	Get(v1.BuildID) (*v1.Build, error)
	Logs(id v1.BuildID, path tree.Path, sidecar *string, options *v1.ContainerLogOptions) (io.ReadCloser, error)
}

type SystemDeployBackend interface {
	CreateFromBuild(v1.BuildID) (*v1.Deploy, error)
	CreateFromPath(tree.Path) (*v1.Deploy, error)
	CreateFromVersion(v1.Version) (*v1.Deploy, error)
	List() ([]v1.Deploy, error)
	Get(v1.DeployID) (*v1.Deploy, error)
}

type SystemJobBackend interface {
	Run(
		path tree.Path,
		command []string,
		environment definitionv1.ContainerExecEnvironment,
		numRetries *int32,
	) (*v1.Job, error)
	List() ([]v1.Job, error)
	Get(v1.JobID) (*v1.Job, error)
	Runs(v1.JobID) SystemJobRunBackend
}

type SystemJobRunBackend interface {
	List() ([]v1.JobRun, error)
	Get(v1.JobRunID) (*v1.JobRun, error)
	Logs(id v1.JobRunID, sidecar *string, options *v1.ContainerLogOptions) (io.ReadCloser, error)
}

type SystemNodePoolBackend interface {
	List() ([]v1.NodePool, error)
	Get(path tree.PathSubcomponent) (*v1.NodePool, error)
}

type SystemSecretBackend interface {
	List() ([]v1.Secret, error)
	Get(tree.PathSubcomponent) (*v1.Secret, error)
	Set(path tree.PathSubcomponent, value string) error
	Unset(tree.PathSubcomponent) error
}

type SystemServiceBackend interface {
	List() ([]v1.Service, error)
	Get(v1.ServiceID) (*v1.Service, error)
	GetByPath(tree.Path) (*v1.Service, error)
	Logs(
		id v1.ServiceID,
		sidecar *string,
		instance string,
		options *v1.ContainerLogOptions,
	) (io.ReadCloser, error)
}

type SystemTeardownBackend interface {
	Create() (*v1.Teardown, error)
	List() ([]v1.Teardown, error)
	Get(v1.TeardownID) (*v1.Teardown, error)
}
