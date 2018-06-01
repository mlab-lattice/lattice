package v1

import (
	"io"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type Interface interface {
	Systems() SystemClient
}

type SystemClient interface {
	Create(id v1.SystemID, definitionURL string) (*v1.System, error)
	List() ([]v1.System, error)
	Get(v1.SystemID) (*v1.System, error)
	Delete(v1.SystemID) error

	Versions(v1.SystemID) ([]v1.SystemVersion, error)
	Builds(v1.SystemID) BuildClient
	Deploys(v1.SystemID) DeployClient
	Teardowns(v1.SystemID) TeardownClient
	Services(v1.SystemID) ServiceClient
	Secrets(v1.SystemID) SecretClient
}

type BuildClient interface {
	Create(version v1.SystemVersion) (*v1.Build, error)
	List() ([]v1.Build, error)
	Get(v1.BuildID) (*v1.Build, error)
	Logs(id v1.BuildID, path tree.NodePath, component string, follow bool) (io.ReadCloser, error)
}

type DeployClient interface {
	CreateFromBuild(v1.BuildID) (*v1.Deploy, error)
	CreateFromVersion(v1.SystemVersion) (*v1.Deploy, error)
	List() ([]v1.Deploy, error)
	Get(v1.DeployID) (*v1.Deploy, error)
}

type TeardownClient interface {
	Create() (*v1.Teardown, error)
	List() ([]v1.Teardown, error)
	Get(v1.TeardownID) (*v1.Teardown, error)
}

type ServiceClient interface {
	List() ([]v1.Service, error)
	Get(id v1.ServiceID) (*v1.Service, error)
	GetByServicePath(path tree.NodePath) (*v1.Service, error)
	Logs(id v1.ServiceID, component string, instance string, follow bool) (io.ReadCloser, error)
}

type SecretClient interface {
	List() ([]v1.Secret, error)
	Get(path tree.NodePath, name string) (*v1.Secret, error)
	Set(path tree.NodePath, name, value string) error
	Unset(path tree.NodePath, name string) error
}
