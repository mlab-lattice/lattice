package v1

import (
	"io"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type Interface interface {
	Systems() SystemClient
}

type SystemClient interface {
	Create(id v1.SystemID, definitionURL string) (*v1.System, error)
	List() ([]v1.System, error)
	Get(v1.SystemID) (*v1.System, error)
	Delete(v1.SystemID) error

	Versions(v1.SystemID) ([]v1.Version, error)
	Builds(v1.SystemID) SystemBuildClient
	Deploys(v1.SystemID) SystemDeployClient
	Teardowns(v1.SystemID) SystemTeardownClient
	Services(v1.SystemID) SystemServiceClient
	Jobs(v1.SystemID) SystemJobClient
	Secrets(v1.SystemID) SystemSecretClient
}

type SystemBuildClient interface {
	CreateFromVersion(v1.Version) (*v1.Build, error)
	CreateFromPath(path tree.Path) (*v1.Build, error)
	List() ([]v1.Build, error)
	Get(v1.BuildID) (*v1.Build, error)
	Logs(id v1.BuildID, path tree.Path, sidecar *string, options *v1.ContainerLogOptions) (io.ReadCloser, error)
}

type SystemDeployClient interface {
	CreateFromBuild(v1.BuildID) (*v1.Deploy, error)
	CreateFromPath(tree.Path) (*v1.Deploy, error)
	CreateFromVersion(v1.Version) (*v1.Deploy, error)
	List() ([]v1.Deploy, error)
	Get(v1.DeployID) (*v1.Deploy, error)
}

type SystemTeardownClient interface {
	Create() (*v1.Teardown, error)
	List() ([]v1.Teardown, error)
	Get(v1.TeardownID) (*v1.Teardown, error)
}

type SystemServiceClient interface {
	List() ([]v1.Service, error)
	Get(id v1.ServiceID) (*v1.Service, error)
	GetByServicePath(path tree.Path) (*v1.Service, error)
	Logs(id v1.ServiceID, sidecar, instance *string, options *v1.ContainerLogOptions) (io.ReadCloser, error)
}

type SystemJobClient interface {
	Create(path tree.Path, command []string, environment definitionv1.ContainerEnvironment) (*v1.Job, error)
	List() ([]v1.Job, error)
	Get(v1.JobID) (*v1.Job, error)
	Logs(id v1.JobID, sidecar *string, options *v1.ContainerLogOptions) (io.ReadCloser, error)
}

type SystemSecretClient interface {
	List() ([]v1.Secret, error)
	Get(path tree.PathSubcomponent) (*v1.Secret, error)
	Set(path tree.PathSubcomponent, value string) error
	Unset(path tree.PathSubcomponent) error
}
