package v1

import (
	"io"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type Interface interface {
	// System
	CreateSystem(systemID v1.SystemID, definitionURL string) (*v1.System, error)
	ListSystems() ([]v1.System, error)
	GetSystem(v1.SystemID) (*v1.System, error)
	DeleteSystem(v1.SystemID) error

	// Build
	Build(systemID v1.SystemID, definitionRoot tree.Node, v v1.SystemVersion) (*v1.Build, error)
	ListBuilds(v1.SystemID) ([]v1.Build, error)
	GetBuild(v1.SystemID, v1.BuildID) (*v1.Build, error)
	BuildLogs(systemID v1.SystemID, buildID v1.BuildID, path tree.NodePath, component string,
		logOptions *v1.ContainerLogOptions) (io.ReadCloser, error)

	// Deploy
	DeployBuild(v1.SystemID, v1.BuildID) (*v1.Deploy, error)
	DeployVersion(systemID v1.SystemID, definitionRoot tree.Node, version v1.SystemVersion) (*v1.Deploy, error)
	ListDeploys(v1.SystemID) ([]v1.Deploy, error)
	GetDeploy(v1.SystemID, v1.DeployID) (*v1.Deploy, error)

	// Teardown
	TearDown(v1.SystemID) (*v1.Teardown, error)
	ListTeardowns(v1.SystemID) ([]v1.Teardown, error)
	GetTeardown(v1.SystemID, v1.TeardownID) (*v1.Teardown, error)

	// Service
	ListServices(v1.SystemID) ([]v1.Service, error)
	GetService(v1.SystemID, v1.ServiceID) (*v1.Service, error)
	GetServiceByPath(v1.SystemID, tree.NodePath) (*v1.Service, error)
	ServiceLogs(systemID v1.SystemID, serviceID v1.ServiceID, component string,
		instance string, logOptions *v1.ContainerLogOptions) (io.ReadCloser, error)

	// System Secret
	ListSystemSecrets(v1.SystemID) ([]v1.Secret, error)
	GetSystemSecret(systemID v1.SystemID, path tree.NodePath, name string) (*v1.Secret, error)
	SetSystemSecret(systemID v1.SystemID, path tree.NodePath, name, value string) error
	UnsetSystemSecret(systemID v1.SystemID, path tree.NodePath, name string) error

	ListNodePools(v1.SystemID) ([]v1.NodePool, error)
	GetNodePool(v1.SystemID, v1.NodePoolPath) (*v1.NodePool, error)
}
